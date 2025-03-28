package http

import (
	"github.com/alexperezortuno/go-simple-auth/infra"
	"github.com/alexperezortuno/go-simple-auth/internal/errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// Endpoint para renovar un token existente
func RenewToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	expirationTime := time.Now().Add(time.Hour)
	err := infra.RenewToken(tokenString, expirationTime)

	if err != nil {
		c.JSON(http.StatusUnauthorized, errors.NewCustomError(errors.TokenInvalidOrExpired))
		return
	}

	// Extraer el username del token
	username, err := infra.ExtractUsernameFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errors.NewCustomError(errors.InvalidToken))
		return
	}

	// Generar un nuevo token
	newToken, err := infra.GenerateToken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewCustomError(errors.FailedToGenerateToken))
		return
	}

	// Eliminar el token viejo y almacenar el nuevo
	infra.DeleteToken(tokenString)

	c.JSON(http.StatusOK, gin.H{"new_token": newToken})
}

func ValidateToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	_, err := infra.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errors.NewCustomError(errors.InvalidToken))
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}
