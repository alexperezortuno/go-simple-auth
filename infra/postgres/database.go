package postgres

import (
	"fmt"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
)

var Db *gorm.DB

func connect() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Initialize(conf config.Config) {
	Db, _ = gorm.Open(postgres.Open(dsn(conf)), &gorm.Config{})

	// Configurar el pool
	sqlDB, err := Db.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(conf.DBPoolMaxConns)
	sqlDB.SetMaxIdleConns(conf.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(conf.DBConnMaxLifetime) * time.Minute)

	// Verificar conexión
	err = sqlDB.Ping()
	if err != nil {
		log.Fatal(err)
	}

	sqlDB.Stats()

	log.Println("connection has been successfully")
}

func CloseConnection() error {
	sqlDB, err := Db.DB()

	if err != nil {
		return err
	}

	err = sqlDB.Close()

	if err != nil {
		return err
	}

	log.Println("Db close")
	return nil
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
		err := Db.AutoMigrate(&User{})
		if err != nil {
			log.Fatal("error migrating user table:", err)
			return
		}
	}
}

// Función para registrar un usuario desde la terminal
func CreateUser(username, password string, conf config.Config) {
	Initialize(conf)
	defer CloseConnection()

	// Hashear la contraseña antes de guardarla
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("error hashing the password:", err)
	}

	// Guardar usuario en la base de datos
	user := User{Username: username, Password: string(hashedPassword)}
	if err := Db.Create(&user).Error; err != nil {
		log.Fatal("error creating user:", err)
	}
	log.Println("user created successfully")
}

// Inicializar la base de datos y crear tabla si no existe
func InitDatabase(conf config.Config) {
	//var err error
	//db, err = gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	//if err != nil {
	//	log.Fatal("failed to connect to the database:", err)
	//}
	//if migrate {
	//	err = db.AutoMigrate(&User{})
	//	if err != nil {
	//		log.Fatal("failed to migrate the database:", err)
	//		return
	//	}
	//}
	Initialize(conf)
	Migrate(conf)
}
