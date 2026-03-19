package handler

import (
	"auth-service/internal/model"
	"auth-service/internal/service"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.LoginRequest true "Login Request"
// @Success      200  {object}  model.LoginResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// Register godoc
// @Summary      Register user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user body model.User true "User Data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := h.svc.Register(&user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Username or email already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user_id": fmt.Sprint(user.ID),
	})
}
