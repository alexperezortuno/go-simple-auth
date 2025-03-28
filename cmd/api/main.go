package main

import (
	"fmt"
	"github.com/alexperezortuno/go-simple-auth/cmd/api/bootstrap"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"github.com/alexperezortuno/go-simple-auth/internal/errors"
	"github.com/alexperezortuno/go-simple-auth/middleware"
	"github.com/gin-contrib/gzip"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

/*
func main() {
	generalConfig()
	handleShutdown()

	// Inicializar base de datos
	migrate := config.GetEnvBool("MIGRATE", false)
	compressed := config.GetEnvBool("COMPRESSED", false)
	dbType := config.GetEnvStr("DB_TYPE", "postgres")
	initDatabase(migrate, dbType)

	// Modo para crear usuario desde terminal
	if len(os.Args) >= 3 && os.Args[1] == "create-user" {
		m := config.GetEnvBool(os.Args[4], false)
		log.Printf("creating user %s with password %s migrate is active %s", os.Args[2], os.Args[3], os.Args[4])
		if m {
			initDatabase(m, dbType)
		}
		createUser(os.Args[2], os.Args[3], m)
		return
	}

	r := gin.Default()

	if compressed {
		r.Use(gzip.Gzip(gzip.BestCompression))
	}

	port := config.GetEnvInt("PORT", 8080)
	log.Printf("server running in http://localhost:%d", port)
	err := r.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("failed to start server:", err)
		return
	}
}
*/

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatal(err)
	}
}
