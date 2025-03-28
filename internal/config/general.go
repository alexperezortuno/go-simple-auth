package config

import (
	"github.com/gin-gonic/gin"
	"log"
)

type Config struct {
	Host             string
	Port             int
	DBHost           string
	DBUser           string
	DBPass           string
	DBName           string
	DBPort           int
	DBEngine         string
	DBMigrate        bool
	SSLMode          string
	Schema           string
	UrlTokenVerifier string
	SubDir           string
	ShutdownTimeout  int
}

func GetConfigs() Config {
	return Config{
		Host:             GetEnvStr("HOST", ""),
		Port:             GetEnvInt("PORT", 8080),
		SubDir:           GetEnvStr("CONTEXT", "api"),
		DBHost:           GetEnvStr("DB_HOST", "localhost"),
		DBUser:           GetEnvStr("DB_USER", "postgres"),
		DBPass:           GetEnvStr("DB_PASS", "yourpass"),
		DBName:           GetEnvStr("DB_NAME", "yourdb"),
		DBPort:           GetEnvInt("DB_PORT", 5432),
		DBEngine:         GetEnvStr("DB_ENGINE", "postgres"),
		DBMigrate:        GetEnvBool("DB_MIGRATE", false),
		SSLMode:          GetEnvStr("SSL_MODE", "disable"),
		Schema:           GetEnvStr("DB_SCHEMA", "public"),
		UrlTokenVerifier: GetEnvStr("URL_TOKEN_VERIFIER", "localhost"),
		ShutdownTimeout:  GetEnvInt("SHUTDOWN_TIMEOUT", 10),
	}
}

func SetGinMode() {
	release := GetEnvStr("RELEASE", "prod")

	switch release {
	case "dev":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "prod":
		gin.SetMode(gin.ReleaseMode)
	default:
		log.Fatalf("Invalid environment: %s", release)
	}
}
