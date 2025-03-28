package postgres

import (
	"fmt"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var db *gorm.DB

func connect() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

var Connection *gorm.DB

func Initialize(conf config.Config) {
	Connection, _ = gorm.Open(postgres.Open(dsn(conf)), &gorm.Config{})
	log.Println("connection has been successfully")
}

func CloseConnection() {
	sqlDB, err := Connection.DB()

	if err != nil {
		log.Fatal(err)
	}

	err = sqlDB.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connection close")
}

func dsn(conf config.Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		conf.DBHost,
		conf.DBUser,
		conf.DBPass,
		conf.DBName,
		conf.DBPort,
		conf.SSLMode)
}

func Migrate(conf config.Config) {
	if conf.DBMigrate {
		err := Connection.AutoMigrate(&User{})
		if err != nil {
			log.Fatal("error migrating user table:", err)
			return
		}
	}
}
