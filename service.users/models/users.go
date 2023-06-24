package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Users struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Username      string    `json:"username" unique:"true" validate:"required,min=3,max=50"`
	Password      string    `json:"password" validate:"required,min=8,max=64"`
	FirstName     string    `json:"firstName" validate:"required,min=2,max=50"`
	LastName      string    `json:"lastName" validate:"required,min=2,max=50"`
	Role          UserRole  `json:"role" validate:"required" default:"user"`
	Email         string    `json:"email" validate:"required,email" unique:"true"`
	EmailVerified bool      `json:"emailVerified" default:"false"`
	AccountActive bool      `json:"accountActive" default:"true"`
	UserCreatedAt time.Time `json:"userCreatedAt" default:"now"`
}

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleManager UserRole = "manager"
	UserRoleUser    UserRole = "user"
)

func MigrateUsers(db *gorm.DB) error {
	err := db.AutoMigrate(&Users{})
	return err
}
