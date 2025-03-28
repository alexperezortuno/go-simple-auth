package postgres

import (
	"context"
	"github.com/alexperezortuno/go-simple-auth/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db}
}

// Función para registrar un usuario desde la terminal
func (r *UserRepository) Create(ctx context.Context, username, password string) error {
	// Hashear la contraseña antes de guardarla
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("error hashing the password:", err)
		return err
	}

	// Guardar usuario en la base de datos
	user := User{Username: username, Password: string(hashedPassword)}
	if err := db.WithContext(ctx).Create(&user).Error; err != nil {
		log.Fatal("error creating user:", err)
		return err
	}
	log.Println("user created successfully")
	return nil
}

func (r *UserRepository) Get(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	if err := db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
