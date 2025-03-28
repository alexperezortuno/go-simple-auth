package middleware

import (
	"fmt"
	"github.com/alexperezortuno/go-simple-auth/internal/errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

var limiter = rate.NewLimiter(1, 3)

func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("PANIC recovered: %v\nStack: %s\n", err, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewCustomError(errors.InternalServerError))
			}
		}()
		c.Next()
	}
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errors.NewCustomError(errors.RateLimitExceeded))
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
