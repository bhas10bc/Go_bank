package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JSONResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func NewJSONResponse(status int, message string, data any) JSONResponse {
	return JSONResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// Helper to send JSON responses using Fiber
func sendJSONResponse(c *fiber.Ctx, response JSONResponse, statusCode int) error {
	return c.Status(statusCode).JSON(response)
}

type APIServer struct {
	listenAddr  string
	Store       Storage
	JWTSecret   []byte
	accountChan chan CreateAccountRequest
	wg          sync.WaitGroup
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func NewApiServer(listenAddr string, store Storage, jwtSecret string) *APIServer {
	return &APIServer{
		listenAddr:  listenAddr,
		Store:       store,
		JWTSecret:   []byte(jwtSecret),
		accountChan: make(chan CreateAccountRequest),
	}
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Number    string    `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Number:    "+91" + strconv.Itoa(rand.Intn(100000)),
		CreatedAt: time.Now().UTC(),
	}
}

// Define a custom claims struct
type Claims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

// JWTMiddleware checks for a valid JWT token
func (s *APIServer) JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return sendJSONResponse(c, NewJSONResponse(http.StatusUnauthorized, "Missing authorization header", nil), http.StatusUnauthorized)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.JWTSecret, nil
		})

		if err != nil {
			return sendJSONResponse(c, NewJSONResponse(http.StatusUnauthorized, "Invalid or expired token", nil), http.StatusUnauthorized)
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			c.Locals("userId", claims.UserID)
			return c.Next()
		}

		return sendJSONResponse(c, NewJSONResponse(http.StatusUnauthorized, "Invalid token claims", nil), http.StatusUnauthorized)
	}
}

// StartQueue starts a goroutine to process the account creation queue
func (s *APIServer) StartQueue() {
	go func() {
		for accountReq := range s.accountChan {
			s.wg.Add(1)
			go func(req CreateAccountRequest) {
				defer s.wg.Done()
				account := NewAccount(req.FirstName, req.LastName)
				if err := s.Store.CreateAccount(account); err != nil {
					log.Printf("Error creating account: %v", err)
				}
			}(accountReq)
		}
	}()
}

func (s *APIServer) Run() {
	app := fiber.New()

	app.Get("/ping", func(c *fiber.Ctx) error {
		return sendJSONResponse(c, NewJSONResponse(http.StatusOK, "vanakkam da mapla", nil), http.StatusOK)
	})

	// app.Post("/login", s.handleLogin)

	// protected := app.Group("/")
	// protected.Use(s.JWTMiddleware())

	app.Post("/get-account", s.handleGetAccount)
	app.Post("/get-accounts", s.handleGetAccounts)
	app.Post("/create-account", s.handleCreateAccount)
	app.Delete("/delete-account", s.handleDeleteAccount)

	s.StartQueue()

	log.Printf("Server running on %s", s.listenAddr)
	log.Fatal(app.Listen(s.listenAddr))
}

type AccountRequest struct {
	ID int `json:"id"`
}

func (s *APIServer) handleGetAccount(c *fiber.Ctx) error {
	var req AccountRequest
	if err := c.BodyParser(&req); err != nil {
		return sendJSONResponse(c, NewJSONResponse(http.StatusBadRequest, "Invalid request payload", nil), http.StatusBadRequest)
	}

	account, err := s.Store.GetAccount(req.ID)
	if err != nil {
		return sendJSONResponse(c, NewJSONResponse(http.StatusNotFound, "Account not found", nil), http.StatusNotFound)
	}

	return sendJSONResponse(c, NewJSONResponse(http.StatusOK, "Account retrieved successfully", account), http.StatusOK)
}

func (s *APIServer) handleCreateAccount(c *fiber.Ctx) error {
	var createAccountReq CreateAccountRequest

	if err := c.BodyParser(&createAccountReq); err != nil {
		return sendJSONResponse(c, NewJSONResponse(http.StatusBadRequest, "Invalid request payload", nil), http.StatusBadRequest)
	}

	s.accountChan <- createAccountReq
	return sendJSONResponse(c, NewJSONResponse(http.StatusAccepted, "Account creation in progress", nil), http.StatusAccepted)
}

func (s *APIServer) handleDeleteAccount(c *fiber.Ctx) error {
	var delReq AccountRequest

	if err := c.BodyParser(&delReq); err != nil {
		return sendJSONResponse(c, NewJSONResponse(http.StatusBadRequest, "Invalid request payload", nil), http.StatusBadRequest)
	}

	if err := s.Store.DeleteAccount(delReq.ID); err != nil {
		return sendJSONResponse(c, NewJSONResponse(http.StatusNotFound, "Account not found", nil), http.StatusNotFound)
	}

	return sendJSONResponse(c, NewJSONResponse(http.StatusOK, "Account deleted successfully", nil), http.StatusOK)
}

func (s *APIServer) handleGetAccounts(c *fiber.Ctx) error {
	accounts, err := s.Store.GetAccounts()
	if err != nil {
		return sendJSONResponse(c, NewJSONResponse(http.StatusInternalServerError, "Failed to retrieve accounts", nil), http.StatusInternalServerError)
	}
	return sendJSONResponse(c, NewJSONResponse(http.StatusOK, "Accounts retrieved successfully", accounts), http.StatusOK)
}

// GenerateToken creates a new JWT token for a user
func (s *APIServer) GenerateToken(userID int) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JWTSecret)
}

func (s *APIServer) handleLogin(c *fiber.Ctx) error {
	// TODO: Implement user authentication logic here
	userID := 123 // This should be the authenticated user's ID
	token, err := s.GenerateToken(userID)
	if err != nil {
		return sendJSONResponse(c, NewJSONResponse(http.StatusInternalServerError, "Failed to generate token", nil), http.StatusInternalServerError)
	}

	return sendJSONResponse(c, NewJSONResponse(http.StatusOK, "Login successful", fiber.Map{"token": token}), http.StatusOK)
}