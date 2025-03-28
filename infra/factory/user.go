package factory

import (
	"github.com/alexperezortuno/go-simple-auth/infra/postgres"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"github.com/alexperezortuno/go-simple-auth/usecase"
	"gorm.io/gorm"
)

func NewUserRepository(conf config.Config, db *gorm.DB) usecase.UserRepository {
	switch conf.DBEngine {
	case "postgres":
		return &postgres.PostgresUserRepository{Db: db}
	// Add cases for other databases here
	default:
		return nil
	}
}
