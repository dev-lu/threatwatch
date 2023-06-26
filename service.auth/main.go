package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/golang-jwt/jwt/v5"
	"github.com/segmentio/kafka-go"

	jwtware "github.com/gofiber/contrib/jwt"
)

func writeToKafka(message string) {
	conn, err := kafka.DialLeader(context.Background(), "tcp", "kafka:9092", "auth_topic", 0)
	if err != nil {
		fmt.Println("Error to connect to Kafka: ", err)
		return
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(time.Second * 10))

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"kafka:9092"},
		Topic:    "auth_topic",
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Value: []byte(message),
	})
	if err != nil {
		fmt.Println("Error writing message to Kafka:", err)
	}
}

var (
	// Do not do this in production.
	// In production, generate the private key and public key pair in advance
	privateKey *rsa.PrivateKey
)

type Claims struct {
	Issuer       string `json:"iss"`
	Subject      string `json:"sub"`
	EmailAddress string `json:"email"`
	Role         string `json:"role"`
	IssuedAt     int    `json:"iat"`
	ExpiresAt    int    `json:"exp"`
}

func getCredentials(email string) (map[string]interface{}, error) {
	resp, err := http.Get("http://users_service:4000/api/v1/get_credentials/" + email)
	if err != nil {
		errorMessage := "Cannot get credentials: " + err.Error()
		writeToKafka(errorMessage)
		fmt.Println(errorMessage)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body into a byte slice
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorMessage := "Cannot read response Body: " + err.Error()
		writeToKafka(errorMessage)
		fmt.Println(errorMessage)
		return nil, err
	}

	// Unmarshal the response body into a Go map
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		errorMessage := "Cannot unmarshal response body: " + err.Error()
		writeToKafka(errorMessage)
		fmt.Println(errorMessage)
		return nil, err
	}

	return data, nil
}

func login(c *fiber.Ctx) error {
	email := c.FormValue("email")
	pass := c.FormValue("pass")

	creds, err := getCredentials(email)
	if err != nil {
		errorMessage := "Cannot get credentials: " + err.Error()
		writeToKafka(errorMessage)
		fmt.Println(errorMessage)
		return err
	}

	if email != creds["email"] || pass != creds["password"] {
		c.SendStatus(fiber.StatusUnauthorized)
		errorMessage := "Invalid username or password: " + email
		writeToKafka(errorMessage)
		fmt.Println(errorMessage)
		return c.JSON(fiber.Map{
			"message": "invalid username or password",
		})
	}

	claims := jwt.MapClaims{
		"iss":   "threatwatch-auth-service",
		"sub":   creds["username"],
		"email": creds["email"],
		"role":  creds["role"],
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t, err := token.SignedString(privateKey)
	if err != nil {
		errorMessage := "Cannot create jwt token: " + err.Error()
		writeToKafka(errorMessage)
		fmt.Println(errorMessage)
		log.Printf("token.SignedString: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    t,
		Expires:  time.Now().Add(time.Hour * 72),
		HTTPOnly: true,
	}

	successMessage := "Successful login: " + claims["sub"].(string)
	writeToKafka(successMessage)
	fmt.Println(successMessage)
	c.Cookie(&cookie)
	return c.JSON(fiber.Map{"message": "success"})
}

func validate(c *fiber.Ctx) error {
	tokenString := c.Cookies("jwt")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return privateKey.Public(), nil
	})

	if err != nil {
		errorMessage := "Cannot validate jwt: " + err.Error()
		writeToKafka(errorMessage)
		fmt.Println(errorMessage)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !token.Valid {
		errorMessage := "Invalid jwt: " + tokenString
		writeToKafka(errorMessage)
		fmt.Println(errorMessage)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.JSON(fiber.Map{"message": "success"})
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	successMessage := "successfully logged out: " + c.Cookies("jwt")
	writeToKafka(successMessage)
	fmt.Println(successMessage)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func restricted(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["sub"].(string)
	return c.SendString("Welcome " + name)
}

func main() {

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	// Generate a new private/public key pair on each run. See note above.
	rng := rand.Reader
	var err error
	privateKey, err = rsa.GenerateKey(rng, 2048)
	if err != nil {
		errorMessage := "Error creating key pair: " + err.Error()
		writeToKafka(errorMessage)
		fmt.Println(errorMessage)
	}

	// Unauthenticated routes
	app.Post("/login", login)
	app.Get("/api/v1/auth/health", healthCheck)
	app.Get("/api/v1/auth/validate", validate)
	app.Get("/api/v1/auth/.well-known/jwks.json", validate)

	// JWT Middleware (everything below needs authentication)
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{
			JWTAlg: jwtware.RS256,
			Key:    privateKey.Public(),
		},
		TokenLookup: "cookie:jwt",
	}))

	// Restricted Routes
	app.Get("/restricted", restricted)
	app.Post("/logout", Logout)

	// Users service
	app.Get("/api/v1/users/health", func(c *fiber.Ctx) error {
		url := "http://users_service:4000/api/v1/users_service/health"
		if err := proxy.Do(c, url); err != nil {
			return err
		}
		c.Response().Header.Del(fiber.HeaderServer)
		return nil
	})

	app.Listen(":4000")
}
