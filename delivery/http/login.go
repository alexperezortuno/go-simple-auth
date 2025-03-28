package http

import (
	"fmt"
	"github.com/alexperezortuno/go-simple-auth/infra"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"github.com/alexperezortuno/go-simple-auth/internal/errors"
	"github.com/alexperezortuno/go-simple-auth/usecase"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type LoginHandler struct {
	service usecase.UserService
}

func NewLoginHandler(e *gin.Engine, service usecase.UserService, conf config.Config) {
	handler := &LoginHandler{service}
	e.POST(fmt.Sprintf("%s/login", conf.SubDir), handler.Login)
}

// Endpoint para autenticarse y generar un token
func (s *LoginHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewCustomError(errors.InvalidFormat))
		return
	}

	// Buscar usuario en la base de datos
	user, err := s.service.GetUser(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errors.NewCustomError(errors.InvalidUserOrPassword))
		return
	}

	// Comparar la contraseña hasheada
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, errors.NewCustomError(errors.InvalidUserOrPassword))
		return
	}

	// Generar token
	token, err := infra.GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewCustomError(errors.FailedToGenerateToken))
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
