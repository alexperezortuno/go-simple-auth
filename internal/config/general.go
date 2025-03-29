package config

import (
	"github.com/gin-gonic/gin"
	"log"
)

type Config struct {
	Host              string
	Port              int
	DBHost            string
	DBUser            string
	DBPass            string
	DBName            string
	DBPort            int
	DBEngine          string
	DBMigrate         bool
	DBPoolMaxConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime int
	SSLMode           string
	Schema            string
	UrlTokenVerifier  string
	SubDir            string
	ShutdownTimeout   int
	Release           string
}

func GetConfigs() Config {
	return Config{
		Host:              GetEnvStr("HOST", ""),
		Port:              GetEnvInt("PORT", 8080),
		SubDir:            GetEnvStr("CONTEXT", "api"),
		DBHost:            GetEnvStr("DB_HOST", "localhost"),
		DBUser:            GetEnvStr("DB_USER", "postgres"),
		DBPass:            GetEnvStr("DB_PASS", "yourpass"),
		DBName:            GetEnvStr("DB_NAME", "yourdb"),
		DBPort:            GetEnvInt("DB_PORT", 5432),
		DBEngine:          GetEnvStr("DB_ENGINE", "postgres"),
		DBMigrate:         GetEnvBool("DB_MIGRATE", false),
		DBPoolMaxConns:    GetEnvInt("DB_POOL_MAX_CONNS", 10),
		DBMaxIdleConns:    GetEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: GetEnvInt("DB_CONN_MAX_LIFETIME", 5),
		SSLMode:           GetEnvStr("SSL_MODE", "disable"),
		Schema:            GetEnvStr("DB_SCHEMA", "public"),
		UrlTokenVerifier:  GetEnvStr("URL_TOKEN_VERIFIER", "localhost"),
		ShutdownTimeout:   GetEnvInt("SHUTDOWN_TIMEOUT", 10),
		Release:           GetEnvStr("RELEASE", "prod"),
	}
}

func (c *Config) SetGinMode() {
	switch c.Release {
	case "dev":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "prod":
		gin.SetMode(gin.ReleaseMode)
	default:
		log.Fatalf("Invalid environment: %s", c.Release)
	}
}
