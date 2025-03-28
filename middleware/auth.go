package middleware

import (
	"github.com/alexperezortuno/go-simple-auth/internal/errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

// Middleware para verificar tokens
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, &errors.CustomError{
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
			c.JSON(http.StatusUnauthorized, &errors.CustomError{
				Message: "token invalid or expired",
				Code:    -1002,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
