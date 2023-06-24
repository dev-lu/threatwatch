package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

func checkUsersServiceHealth() string {
	resp, err := http.Get("http://users_service:4000/api/v1/users_service/health")
	if err != nil {
		log.Printf("Error checking users service health: %v", err)
		return "Error"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Users service health check failed with status code %d", resp.StatusCode)
		return "DOWN"
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return "Error"
		}

		var responseMap map[string]interface{}
		err = json.Unmarshal(body, &responseMap)
		if err != nil {
			log.Printf("Error parsing response body: %v", err)
			return "Error"
		}

		status, ok := responseMap["Status"].(string)
		if !ok {
			log.Printf("Error parsing response body: Message is not a string")
			return "Error"
		}

		return status
	}
}

func checkAuthServiceHealth() string {
	resp, err := http.Get("http://auth_service:4000/api/v1/auth/health")
	if err != nil {
		log.Printf("Error checking ath service health: %v", err)
		return "Error"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Auth service health check failed with status code %d", resp.StatusCode)
		return "DOWN"
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return "Error"
		}

		var responseMap map[string]interface{}
		err = json.Unmarshal(body, &responseMap)
		if err != nil {
			log.Printf("Error parsing response body: %v", err)
			return "Error"
		}

		status, ok := responseMap["Status"].(string)
		if !ok {
			log.Printf("Error parsing response body: Message is not a string")
			return "Error"
		}

		return status
	}
}

func checkIpv4ServiceHealth() string {
	resp, err := http.Get("http://ipv4_service:4000/api/v1/ipv4/health")
	if err != nil {
		log.Printf("Error checking IPv4 service health: %v", err)
		return "Error"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("IPv4 service health check failed with status code %d", resp.StatusCode)
		return "DOWN"
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return "Error"
		}

		var responseMap map[string]interface{}
		err = json.Unmarshal(body, &responseMap)
		if err != nil {
			log.Printf("Error parsing response body: %v", err)
			return "Error"
		}

		status, ok := responseMap["Status"].(string)
		if !ok {
			log.Printf("Error parsing response body: Message is not a string")
			return "Error"
		}

		return status
	}
}

func main() {
	app := fiber.New()

	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		userServiceStatus := checkUsersServiceHealth()

		return c.JSON(fiber.Map{
			"Healthcheck": fiber.Map{
				"Status":    "UP",
				"Timestamp": time.Now().Unix(),
			},
			"Services": []fiber.Map{
				{
					"Service":   "Users",
					"Status":    userServiceStatus,
					"Timestamp": time.Now().Unix(),
				},
				{
					"Service":   "Auth",
					"Status":    checkAuthServiceHealth(),
					"Timestamp": time.Now().Unix(),
				},
				{
					"Service":   "IPv4",
					"Status":    checkIpv4ServiceHealth(),
					"Timestamp": time.Now().Unix(),
				},
			},
		})
	})

	log.Fatal(app.Listen(":4000"))
}
