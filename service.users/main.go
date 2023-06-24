package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"service.users/db"
	models "service.users/models"
)

type User struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	Password      string    `json:"password"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Role          string    `json:"role"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	AccountActive bool      `json:"accountActive"`
	UserCreatedAt time.Time `json:"userCreatedAt"`
}

type Repository struct {
	// Pointer to *gorm.DB
	DB *gorm.DB
}

func IsSHA256(s string) bool {
	match, _ := regexp.MatchString("^[a-f0-9]{64}$", s)
	return match
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	// UUID version 4
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	fmt.Println("UUID:", id)
	user.ID = id
	user.UserCreatedAt = time.Now()
	return nil
}

// Create new user
func (r *Repository) CreateUser(context *fiber.Ctx) error {
	user := User{}

	err := context.BodyParser(&user)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	if !IsSHA256(user.Password) {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "password needs to be hashed"})
		return nil
	}

	err = r.DB.Create(&user).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create user"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "user created successfully"})
	return nil
}

// Delete user
func (r *Repository) DeleteUser(context *fiber.Ctx) error {
	userModel := models.Users{}
	id := context.Params("id") // Get ID from URL params
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "Please provide an ID",
		})
		return nil
	}

	//Delete user from DB
	result := r.DB.Where("id = ?", id).Delete(&userModel)

	// Check if the id exists in the database
	if result.RowsAffected == 0 {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "User with the provided ID does not exist",
		})
		return nil
	}

	// Handle error
	if err := result.Error; err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete user",
		})
		return err
	} else {
		context.Status(http.StatusOK).JSON(&fiber.Map{
			"message": "user deleted successfully",
		})
	}
	return nil
}

// Get all users
func (r *Repository) GetUsers(context *fiber.Ctx) error {
	userModels := &[]models.Users{}

	err := r.DB.Find(userModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get users"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "users retrieved successfully",
		"data":    userModels,
	})
	return nil
}

// Get user by ID
func (r *Repository) GetUserByID(context *fiber.Ctx) error {

	id := context.Params("id")

	userModel := &models.Users{}
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	err := r.DB.Where("id = ?", id).First(userModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get user"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "user id fetched successfully",
		"data":    userModel,
	})
	return nil
}

// Get credentials
func (r *Repository) GetCredentials(context *fiber.Ctx) error {
	email := context.Params("email")

	userModel := &models.Users{}
	if email == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "provide username or email address",
		})
		return nil
	}
	err := r.DB.Select("id, username, email, password, role").Where("email = ?", email).First(userModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get user"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"id":       userModel.ID,
		"username": userModel.Username,
		"email":    userModel.Email,
		"password": userModel.Password,
		"role":     userModel.Role,
	})
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return u.Password == password
}

// Update user password
func (r *Repository) UpdateUserPassword(context *fiber.Ctx) error {
	req := struct {
		UserID      uuid.UUID `json:"userId" validate:"required"`
		OldPassword string    `json:"oldPassword" validate:"required,min=8,max=64"`
		NewPassword string    `json:"newPassword" validate:"required,min=8,max=64"`
	}{}
	err := context.BodyParser(&req)
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "invalid request body"})
		return err
	}

	if req.UserID == uuid.Nil || req.OldPassword == "" || req.NewPassword == "" {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "missing required fields"})
		return nil
	}

	user := User{}
	err = r.DB.Where("id = ?", req.UserID).First(&user).Error
	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "user not found"})
		return err
	}

	if !user.CheckPassword(req.OldPassword) {
		context.Status(http.StatusUnauthorized).JSON(
			&fiber.Map{"message": "incorrect old password"})
		return nil
	}

	user.Password = req.NewPassword
	err = r.DB.Save(&user).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not update user password"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "user password updated successfully"})
	return nil
}

func healthCheck(context *fiber.Ctx) error {
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"Status":    "UP",
		"Timestamp": time.Now().Unix(),
	})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")
	api.Post("/create_user", r.CreateUser)
	api.Delete("/delete_user/:id", r.DeleteUser)
	api.Get("/get_user/:id", r.GetUserByID)
	api.Get("/get_credentials/:email", r.GetCredentials)
	api.Get("/users", r.GetUsers)
	api.Put("/update_user_password", r.UpdateUserPassword)
	api.Get("/users_service/health", healthCheck)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	config := &db.Config{
		Host:     "users_db",
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("SSL_Mode"),
	}

	db, err := db.NewConnection(config)
	if err != nil {
		log.Fatal("Could not load the Database")
	}

	err = models.MigrateUsers(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":4000")
}
