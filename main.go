package main

import (
	"fmt"
	"github.com/gin-contrib/gzip"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Estructura del usuario en la base de datos
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
}

type CustomError struct {
	Message string
	Code    int
}

// Variables globales
var (
	db        *gorm.DB
	jwtSecret = []byte(getEnvStr("JWT_SECRET", "secret"))
	subDir    = getEnvStr("CONTEXT", "/auth")
	limiter   = rate.NewLimiter(1, 3)
)

func (e *CustomError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// Estructura para manejar tokens en memoria
var tokenStore = struct {
	sync.RWMutex
	tokens map[string]time.Time
}{tokens: make(map[string]time.Time)}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Procesar la solicitud
		c.Next()

		// Calcular el tiempo de respuesta
		duration := time.Since(startTime)
		log.Printf("request %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
	}
}

func getEnvStr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// Inicializar la base de datos y crear tabla si no existe
func initDatabase(migrate bool) {
	var err error
	db, err = gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to the database:", err)
	}
	if migrate {
		err = db.AutoMigrate(&User{})
		if err != nil {
			log.Fatal("failed to migrate the database:", err)
			return
		}
	}
}

// Función para registrar un usuario desde la terminal
func createUser(username, password string, migrate bool) {
	// Hashear la contraseña antes de guardarla
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("error hashing the password:", err)
	}

	// Guardar usuario en la base de datos
	user := User{Username: username, Password: string(hashedPassword)}
	if err := db.Create(&user).Error; err != nil {
		log.Fatal("error creating user:", err)
	}
	log.Println("user created successfully")
}

// Generar y almacenar token en Redis
func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(time.Hour)
	claims := jwt.MapClaims{
		"username": username,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	tokenStore.Lock()
	tokenStore.tokens[tokenString] = expirationTime
	tokenStore.Unlock()

	return tokenString, nil
}

// Middleware para verificar tokens
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, &CustomError{
				Message: "token required",
				Code:    -1001,
			})
			c.Abort()
			return
		}

		tokenStore.RLock()
		expirationTime, exists := tokenStore.tokens[tokenString]
		log.Printf("exists: %v, expirationTime: %v", exists, expirationTime)
		tokenStore.RUnlock()

		if !exists || time.Now().After(expirationTime) {
			c.JSON(http.StatusUnauthorized, &CustomError{
				Message: "token invalid or expired",
				Code:    -1002,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func rateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, CustomError{Message: "rate limit exceeded", Code: -1003})
			return
		}
		c.Next()
	}
}

// Endpoint para autenticarse y generar un token
func loginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &CustomError{
			Message: "invalid format",
			Code:    -4001,
		})
		return
	}

	var user User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, &CustomError{
			Message: "invalid user or password",
			Code:    -4002,
		})
		return
	}

	// Comparar la contraseña hasheada
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, &CustomError{
			Message: "invalid user or password",
			Code:    -4003,
		})
		return
	}

	// Generar token
	token, err := generateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &CustomError{
			Message: "failed to generate token",
			Code:    -4004,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Endpoint para renovar un token existente
func renewTokenHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	tokenStore.RLock()
	expirationTime, exists := tokenStore.tokens[tokenString]
	tokenStore.RUnlock()

	if !exists || time.Now().After(expirationTime) {
		c.JSON(http.StatusUnauthorized, &CustomError{
			Message: "token invalid or expired",
			Code:    -3001,
		})
		return
	}

	// Extraer el username del token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	username, ok := claims["username"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, &CustomError{
			Message: "invalid token",
			Code:    -3002,
		})
		return
	}

	// Generar un nuevo token
	newToken, err := generateToken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &CustomError{
			Message: "failed to generate token",
			Code:    -3003,
		})
		return
	}

	// Eliminar el token viejo y almacenar el nuevo
	tokenStore.Lock()
	delete(tokenStore.tokens, tokenString)
	tokenStore.Unlock()

	c.JSON(http.StatusOK, gin.H{"new_token": newToken})
}

func validateToken(tokenString string) (string, *CustomError) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", &CustomError{
			Message: fmt.Sprintf("invalid token: %s", err.Error()),
			Code:    -2001,
		}
	}

	if !token.Valid {
		return "", &CustomError{
			Message: "invalid token",
			Code:    -2002,
		}
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", &CustomError{
			Message: "invalid token",
			Code:    -2003,
		}
	}

	return username, nil
}

func validateHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	_, err := validateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, &CustomError{
			Message: "invalid token",
			Code:    err.Code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// Función para limpiar la memoria antes de salir
func cleanup() {
	fmt.Println("\n🧹 Cleaning memory before exiting...")
	tokenStore.Lock()
	tokenStore.tokens = make(map[string]time.Time) // Vaciar los tokens
	tokenStore.Unlock()

	fmt.Println("✅ Memory cleaned. Exiting...")
	os.Exit(0) // Salir del programa
}

// Capturar señales del sistema y ejecutar la limpieza
func handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM) // Capturar Ctrl+C o SIGTERM

	go func() {
		<-sigChan // Esperar señal
		cleanup()
	}()
}

func generalConfig() {
	rel := getEnvStr("RELEASE", "prod")
	switch rel {
	case "dev":
		gin.SetMode(gin.DebugMode)
		break
	case "test":
		gin.SetMode(gin.TestMode)
		break
	case "prod":
		gin.SetMode(gin.ReleaseMode)
		break
	default:
		log.Fatalf("Invalid environment: %s", rel)
	}
}

func main() {
	generalConfig()
	handleShutdown()

	// Inicializar base de datos
	migrate := getEnvBool("MIGRATE", false)
	compressed := getEnvBool("COMPRESSED", false)
	initDatabase(migrate)

	// Modo para crear usuario desde terminal
	if len(os.Args) >= 3 && os.Args[1] == "create-user" {
		m := getEnvBool(os.Args[4], false)
		log.Printf("creating user %s with password %s migrate is active %s", os.Args[2], os.Args[3], os.Args[4])
		if m {
			initDatabase(m)
		}
		createUser(os.Args[2], os.Args[3], m)
		return
	}

	r := gin.Default()

	if compressed {
		r.Use(gzip.Gzip(gzip.BestCompression))
	}

	r.Use(gzip.Gzip(gzip.BestSpeed))
	// Añadir el middleware de registro de tiempo de respuesta
	r.Use(requestLogger())
	r.Use(rateLimit())

	// Handlers
	r.POST(fmt.Sprintf("%s/login", subDir), loginHandler)          // Generar token con usuario y contraseña
	r.GET(fmt.Sprintf("%s/health", subDir), func(c *gin.Context) { // Endpoint para verificar la salud del servidor
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Rutas protegidas
	protected := r.Group(subDir)
	protected.Use(authMiddleware())
	protected.POST("/renew", renewTokenHandler)  // Renovar token existente
	protected.POST("/validate", validateHandler) // Validar token

	port := getEnvInt("PORT", 8080)
	log.Printf("server running in http://localhost:%d", port)
	err := r.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("failed to start server:", err)
		return
	}
}
