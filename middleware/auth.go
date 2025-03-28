package middleware

import (
	"github.com/alexperezortuno/go-simple-auth/infra"
	"github.com/alexperezortuno/go-simple-auth/internal/errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Middleware para verificar tokens
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, errors.NewCustomError(errors.TokenRequired))
			c.Abort()
			return
		}

		_, err := infra.ValidateToken(tokenString)

		if err != nil {
			c.JSON(http.StatusUnauthorized, errors.NewCustomError(errors.TokenInvalidOrExpired))
			c.Abort()
			return
		}

		c.Next()
	}
}
