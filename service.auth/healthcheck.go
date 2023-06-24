package main

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

func healthCheck(c *fiber.Ctx) error {
	c.Status(http.StatusOK).JSON(&fiber.Map{
		"Status":    "UP",
		"Timestamp": time.Now().Unix(),
	})
	return nil
}
