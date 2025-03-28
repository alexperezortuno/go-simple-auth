package http

import (
	"fmt"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"github.com/alexperezortuno/go-simple-auth/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"sync"
	"time"
)

// Estructura para manejar tokens en memoria
var tokenStore = struct {
	sync.RWMutex
	tokens map[string]time.Time
}{tokens: make(map[string]time.Time)}

var jwtSecret = []byte(config.GetEnvStr("JWT_SECRET", "secret"))

// Endpoint para renovar un token existente
func RenewToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	tokenStore.RLock()
	expirationTime, exists := tokenStore.tokens[tokenString]
	tokenStore.RUnlock()

	if !exists || time.Now().After(expirationTime) {
		c.JSON(http.StatusUnauthorized, errors.NewCustomError(errors.TokenInvalidOrExpired)&errors.CustomError{
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
		c.JSON(http.StatusUnauthorized, &errors.CustomError{
			Message: "invalid token",
			Code:    -3002,
		})
		return
	}

	// Generar un nuevo token
	newToken, err := generateToken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &errors.CustomError{
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

func validateToken(tokenString string) (string, *errors.CustomError) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", &errors.CustomError{
			Message: fmt.Sprintf("invalid token: %s", err.Error()),
			Code:    -2001,
		}
	}

	if !token.Valid {
		return "", &errors.CustomError{
			Message: "invalid token",
			Code:    -2002,
		}
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", &errors.CustomError{
			Message: "invalid token",
			Code:    -2003,
		}
	}

	return username, nil
}

func ValidateToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	_, err := validateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, &errors.CustomError{
			Message: "invalid token",
			Code:    err.Code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// Generar y almacenar token en memoria
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
