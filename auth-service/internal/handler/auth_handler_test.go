package handler

import (
	"auth-service/internal/model"
	"auth-service/internal/service"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupApp() *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("correlationID", "test-correlation-id")
		return c.Next()
	})
	return app
}

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func newReq(method, path string, body *bytes.Buffer) *http.Request {
	if body == nil {
		body = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestAuthHandler_Register_Success(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Post("/register", NewAuthHandler(mockSvc).Register)

	mockSvc.On("Register", "alice", "password123", "alice@example.com").
		Return(&model.User{Username: "alice", Email: "alice@example.com", Role: "user"}, nil)

	resp, _ := app.Test(newReq("POST", "/register", jsonBody(map[string]string{
		"username": "alice", "password": "password123", "email": "alice@example.com",
	})))

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_MissingFields(t *testing.T) {
	app := setupApp()
	app.Post("/register", NewAuthHandler(new(service.MockAuthService)).Register)

	resp, _ := app.Test(newReq("POST", "/register", jsonBody(map[string]string{"username": "alice"})))

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthHandler_Register_ShortPassword(t *testing.T) {
	app := setupApp()
	app.Post("/register", NewAuthHandler(new(service.MockAuthService)).Register)

	resp, _ := app.Test(newReq("POST", "/register", jsonBody(map[string]string{
		"username": "alice", "password": "short", "email": "a@b.com",
	})))

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthHandler_Register_DuplicateUser(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Post("/register", NewAuthHandler(mockSvc).Register)

	mockSvc.On("Register", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, service.ErrUserExists)

	resp, _ := app.Test(newReq("POST", "/register", jsonBody(map[string]string{
		"username": "alice", "password": "password123", "email": "alice@example.com",
	})))

	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestAuthHandler_Login_Success(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Post("/login", NewAuthHandler(mockSvc).Login)

	mockSvc.On("Login", "alice", "password123").
		Return(&model.TokenPair{AccessToken: "acc", RefreshToken: "ref"}, nil)

	resp, _ := app.Test(newReq("POST", "/login", jsonBody(map[string]string{
		"username": "alice", "password": "password123",
	})))

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	var pair model.TokenPair
	json.NewDecoder(resp.Body).Decode(&pair)
	assert.Equal(t, "acc", pair.AccessToken)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Post("/login", NewAuthHandler(mockSvc).Login)

	mockSvc.On("Login", "alice", "wrong").Return(nil, service.ErrInvalidCredentials)

	resp, _ := app.Test(newReq("POST", "/login", jsonBody(map[string]string{
		"username": "alice", "password": "wrong",
	})))

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	app := setupApp()
	app.Post("/login", NewAuthHandler(new(service.MockAuthService)).Login)

	resp, _ := app.Test(newReq("POST", "/login", bytes.NewBufferString("not-json")))

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ── Refresh ───────────────────────────────────────────────────────────────────

func TestAuthHandler_Refresh_Success(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Post("/refresh", NewAuthHandler(mockSvc).Refresh)

	mockSvc.On("Refresh", "myrefreshtoken").
		Return(&model.TokenPair{AccessToken: "newacc", RefreshToken: "newref"}, nil)

	resp, _ := app.Test(newReq("POST", "/refresh", jsonBody(map[string]string{
		"refresh_token": "myrefreshtoken",
	})))

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Post("/refresh", NewAuthHandler(mockSvc).Refresh)

	mockSvc.On("Refresh", "bad").Return(nil, service.ErrInvalidToken)

	resp, _ := app.Test(newReq("POST", "/refresh", jsonBody(map[string]string{
		"refresh_token": "bad",
	})))

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_Refresh_MissingToken(t *testing.T) {
	app := setupApp()
	app.Post("/refresh", NewAuthHandler(new(service.MockAuthService)).Refresh)

	resp, _ := app.Test(newReq("POST", "/refresh", jsonBody(map[string]string{})))

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ── Logout ────────────────────────────────────────────────────────────────────

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Post("/logout", NewAuthHandler(mockSvc).Logout)

	mockSvc.On("Logout", "mytoken").Return(nil)

	resp, _ := app.Test(newReq("POST", "/logout", jsonBody(map[string]string{
		"refresh_token": "mytoken",
	})))

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Logout_MissingToken(t *testing.T) {
	app := setupApp()
	app.Post("/logout", NewAuthHandler(new(service.MockAuthService)).Logout)

	resp, _ := app.Test(newReq("POST", "/logout", jsonBody(map[string]string{})))

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ── LogoutAll ─────────────────────────────────────────────────────────────────

func TestAuthHandler_LogoutAll_Success(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Post("/logout-all", NewAuthHandler(mockSvc).LogoutAll)

	mockSvc.On("LogoutAll", uint(1)).Return(nil)

	req := newReq("POST", "/logout-all", nil)
	req.Header.Set("X-User-Id", "1")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_LogoutAll_MissingUserID(t *testing.T) {
	app := setupApp()
	app.Post("/logout-all", NewAuthHandler(new(service.MockAuthService)).LogoutAll)

	resp, _ := app.Test(newReq("POST", "/logout-all", nil))

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_LogoutAll_ServiceError(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Post("/logout-all", NewAuthHandler(mockSvc).LogoutAll)

	mockSvc.On("LogoutAll", uint(1)).Return(errors.New("db error"))

	req := newReq("POST", "/logout-all", nil)
	req.Header.Set("X-User-Id", "1")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// ── GetProfile ────────────────────────────────────────────────────────────────

func TestAuthHandler_GetProfile_Success(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Get("/profile", NewAuthHandler(mockSvc).GetProfile)

	user := &model.User{Username: "alice", Email: "alice@example.com", Role: "user"}
	mockSvc.On("GetProfile", uint(1)).Return(user, nil)

	req := newReq("GET", "/profile", nil)
	req.Header.Set("X-User-Id", "1")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "alice", body["username"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_GetProfile_MissingUserID(t *testing.T) {
	app := setupApp()
	app.Get("/profile", NewAuthHandler(new(service.MockAuthService)).GetProfile)

	resp, _ := app.Test(newReq("GET", "/profile", nil))

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_GetProfile_NotFound(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Get("/profile", NewAuthHandler(mockSvc).GetProfile)

	mockSvc.On("GetProfile", uint(99)).Return(nil, errors.New("not found"))

	req := newReq("GET", "/profile", nil)
	req.Header.Set("X-User-Id", "99")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// ── ChangePassword ────────────────────────────────────────────────────────────

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Put("/password", NewAuthHandler(mockSvc).ChangePassword)

	mockSvc.On("ChangePassword", uint(1), "oldpass1", "newpass123").Return(nil)

	req := newReq("PUT", "/password", jsonBody(map[string]string{
		"current_password": "oldpass1", "new_password": "newpass123",
	}))
	req.Header.Set("X-User-Id", "1")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_WrongCurrentPassword(t *testing.T) {
	mockSvc := new(service.MockAuthService)
	app := setupApp()
	app.Put("/password", NewAuthHandler(mockSvc).ChangePassword)

	mockSvc.On("ChangePassword", uint(1), "wrong", "newpass123").Return(service.ErrIncorrectPassword)

	req := newReq("PUT", "/password", jsonBody(map[string]string{
		"current_password": "wrong", "new_password": "newpass123",
	}))
	req.Header.Set("X-User-Id", "1")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_ChangePassword_ShortNewPassword(t *testing.T) {
	app := setupApp()
	app.Put("/password", NewAuthHandler(new(service.MockAuthService)).ChangePassword)

	req := newReq("PUT", "/password", jsonBody(map[string]string{
		"current_password": "oldpass1", "new_password": "short",
	}))
	req.Header.Set("X-User-Id", "1")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthHandler_ChangePassword_MissingUserID(t *testing.T) {
	app := setupApp()
	app.Put("/password", NewAuthHandler(new(service.MockAuthService)).ChangePassword)

	resp, _ := app.Test(newReq("PUT", "/password", jsonBody(map[string]string{
		"current_password": "old", "new_password": "newpass123",
	})))

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_ChangePassword_MissingFields(t *testing.T) {
	app := setupApp()
	app.Put("/password", NewAuthHandler(new(service.MockAuthService)).ChangePassword)

	req := newReq("PUT", "/password", jsonBody(map[string]string{"current_password": "old"}))
	req.Header.Set("X-User-Id", "1")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
