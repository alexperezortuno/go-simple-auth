package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("PANIC recovered: %v\nStack: %s\n", err, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, CustomError{Message: "internal server error", Code: -5000})
			}
		}()
		c.Next()
	}
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, CustomError{Message: "rate limit exceeded", Code: -1003})
			return
		}
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Procesar la solicitud
		c.Next()

		// Calcular el tiempo de respuesta
		duration := time.Since(startTime)
		log.Printf("request %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
	}
}
