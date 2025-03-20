package main

import (
	"fmt"
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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Estructura del usuario en la base de datos
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
}

// Variables globales
var (
	db        *gorm.DB
	jwtSecret = []byte("supersecretkey")
)

type CustomError struct {
	Message string
	Code    int
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// Estructura para manejar tokens en memoria
var tokenStore = struct {
	sync.RWMutex
	tokens map[string]time.Time
}{tokens: make(map[string]time.Time)}

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
func createUser(username, password string) {
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

// Genera un nuevo token JWT con duración de 1 hora
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

	// Guardar el token en memoria con su fecha de expiración
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

// Endpoint para autenticarse y generar un token
func loginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format"})
		return
	}

	var user User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user or password"})
		return
	}

	// Comparar la contraseña hasheada
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user or password"})
		return
	}

	// Generar token
	token, err := generateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalid or expired"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Generar un nuevo token
	newToken, err := generateToken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo renovar el token"})
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
			Code:    http.StatusUnauthorized,
		}
	}

	if !token.Valid {
		return "", &CustomError{
			Message: "invalid token",
			Code:    http.StatusUnauthorized,
		}
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", &CustomError{
			Message: "invalid token",
			Code:    http.StatusUnauthorized,
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
			Code:    http.StatusUnauthorized,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// Función para limpiar la memoria antes de salir
func cleanup() {
	fmt.Println("\n🧹 Limpiando memoria antes de salir...")
	tokenStore.Lock()
	tokenStore.tokens = make(map[string]time.Time) // Vaciar los tokens
	tokenStore.Unlock()

	fmt.Println("✅ Memoria limpiada. Saliendo...")
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

func main() {
	handleShutdown()

	// Inicializar base de datos
	migrate := getEnvBool("MIGRATE", false)
	initDatabase(migrate)

	// Modo para crear usuario desde terminal
	if len(os.Args) >= 3 && os.Args[1] == "create-user" {
		createUser(os.Args[2], os.Args[3])
		return
	}

	r := gin.Default()

	r.POST("/login", loginHandler)          // Generar token con usuario y contraseña
	r.GET("/health", func(c *gin.Context) { // Endpoint para verificar la salud del servidor
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Rutas protegidas
	protected := r.Group("/api")
	protected.Use(authMiddleware())
	protected.POST("/renew", renewTokenHandler)  // Renovar token existente
	protected.POST("/validate", validateHandler) // Validar token
	protected.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Acceso autorizado"})
	})

	port := getEnvInt("PORT", 8080)
	log.Printf("server running in http://localhost:%d", port)
	err := r.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("failed to start server:", err)
		return
	}
}
