package infra

import (
	"errors"
	"fmt"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	e "github.com/alexperezortuno/go-simple-auth/internal/errors"
	"github.com/golang-jwt/jwt/v5"
	"sync"
	"time"
)

// Estructura para manejar tokens en memoria
var tokenStore = struct {
	sync.RWMutex
	tokens map[string]time.Time
}{tokens: make(map[string]time.Time)}

var jwtSecret = []byte(config.GetEnvStr("JWT_SECRET", "secret"))

// Funciones relacionadas con la gestión de tokens
func StoreToken(token string, expirationTime time.Time) {
	tokenStore.Lock()
	tokenStore.tokens[token] = expirationTime
	tokenStore.Unlock()
}

func RenewToken(token string, expirationTime time.Time) error {
	tokenStore.RLock()
	_, exists := tokenStore.tokens[token]
	tokenStore.RUnlock()

	if !exists || time.Now().After(expirationTime) {
		return errors.New(e.GetErrorMessage(e.InvalidUserOrPassword))
	}

	StoreToken(token, expirationTime)
	return nil
}

func DeleteToken(token string) {
	tokenStore.Lock()
	delete(tokenStore.tokens, token)
	tokenStore.Unlock()
}

func IsExpired(token string) bool {
	tokenStore.RLock()
	expirationTime, exists := tokenStore.tokens[token]
	tokenStore.RUnlock()

	return !exists || time.Now().After(expirationTime)
}

func ExtractUsernameFromToken(tokenString string) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", errors.New(e.GetErrorMessage(e.InvalidToken))
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New(e.GetErrorMessage(e.InvalidToken))
	}

	return username, nil
}

// Generar y almacenar token en memoria
func GenerateToken(username string) (string, error) {
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

func CleanupTokens() {
	fmt.Println("\n🧹 Cleaning memory before exiting...")
	tokenStore.Lock()
	tokenStore.tokens = make(map[string]time.Time) // Vaciar los tokens
	tokenStore.Unlock()

	fmt.Println("✅ Memory cleaned. Exiting...")
}

func ValidateToken(tokenString string) (string, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", errors.New(e.GetErrorMessage(e.InvalidToken) + ": " + err.Error())
	}

	if !token.Valid {
		return "", errors.New(e.GetErrorMessage(e.InvalidToken))
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New(e.GetErrorMessage(e.InvalidToken))
	}

	if IsExpired(tokenString) {
		return "", errors.New(e.GetErrorMessage(e.TokenInvalidOrExpired))
	}

	return username, nil
}
