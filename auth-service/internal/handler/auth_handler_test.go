package handler

import (
	"auth-service/internal/model"
	"auth-service/internal/service"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Login(t *testing.T) {
	app := fiber.New()
	mockSvc := new(service.MockAuthService)
	handler := NewAuthHandler(mockSvc)

	app.Post("/login", handler.Login)

	reqBody := model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	mockSvc.On("Login", reqBody).Return(&model.LoginResponse{Token: "test_token"}, nil)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var loginResp model.LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginResp)
	assert.Equal(t, "test_token", loginResp.Token)

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register(t *testing.T) {
	app := fiber.New()
	mockSvc := new(service.MockAuthService)
	handler := NewAuthHandler(mockSvc)

	app.Post("/register", handler.Register)

	user := model.User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}
	jsonBody, _ := json.Marshal(user)

	mockSvc.On("Register", &user).Return(nil)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}
