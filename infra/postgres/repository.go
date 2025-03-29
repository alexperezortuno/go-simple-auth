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

type PostgresUserRepository struct {
	Db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *PostgresUserRepository {
	return &PostgresUserRepository{Db: db}
}

func (r *PostgresUserRepository) Save(ctx context.Context, u *domain.User) error {
	// Hashear la contraseña antes de guardarla
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("error hashing the password:", err)
		return err
	}

	// Guardar usuario en la base de datos
	user := User{Username: u.Username, Password: string(hashedPassword)}
	if err := Db.WithContext(ctx).Create(&user).Error; err != nil {
		log.Fatal("error creating user:", err)
		return err
	}
	log.Println("user created successfully")
	return nil
}

func (r *PostgresUserRepository) Get(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User

	// Check if user exists
	if err := Db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
