package main

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

var servers = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8080",
}

func main() {
	rand.Seed(time.Now().UnixNano())``
	app := fiber.New()

	app.Post("/create-account", proxyRequest)
	app.Post("/get-accounts", proxyRequest)

	// Start the discovery service
	log.Println("Starting discovery service on port 8084...")
	if err := app.Listen(":8084"); err != nil {
		log.Fatal(err)
	}
}

func proxyRequest(c *fiber.Ctx) error {
	// Randomly select one of the backend servers
	target := servers[rand.Intn(len(servers))]
	startTime := time.Now() // Start measuring time

	// Create a new request to forward
	req, err := http.NewRequest(c.Method(), target+c.Path(), bytes.NewReader(c.Body()))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create request")
	}

	// Copy headers from the original request
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Send the request to the selected server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error reaching backend server %s: %v\n", target, err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to reach backend server")
	}
	defer resp.Body.Close()

	// Measure response time
	responseTime := time.Since(startTime)

	// Log the selected server URL and response time
	log.Printf("Selected backend server: %s | Response time: %v\n", target, responseTime)

	// Copy the response headers and status code to the client
	for key, value := range resp.Header {
		c.Set(key, value[0])
	}
	c.Status(resp.StatusCode)

	// Return the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to read response body")
	}
	return c.Send(body)
}
