package handler

import (
	"auth-service/internal/service"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body object{username=string,password=string,email=string} true "Register request"
// @Success      201  {object}  object{message=string,user=object{id=int,username=string,email=string,role=string}}
// @Failure      400  {object}  object{error=string}
// @Failure      409  {object}  object{error=string}
// @Failure      500  {object}  object{error=string}
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username, password and email are required"})
	}
	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password must be at least 8 characters"})
	}

	user, err := h.authService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("register failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	log.Info().Str("correlation_id", correlationID).Uint("user_id", user.ID).Msg("user registered")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "registered successfully",
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"group":    user.Group.Name,
			"roles":    user.Group.RoleNames(),
		},
	})
}

// Login godoc
// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body object{username=string,password=string} true "Login request"
// @Success      200  {object}  object{access_token=string,refresh_token=string}
// @Failure      400  {object}  object{error=string}
// @Failure      401  {object}  object{error=string}
// @Failure      500  {object}  object{error=string}
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	pair, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid username or password"})
		}
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("login failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	log.Info().Str("correlation_id", correlationID).Str("username", req.Username).Msg("user logged in")
	return c.Status(fiber.StatusOK).JSON(pair)
}

// Refresh godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body object{refresh_token=string} true "Refresh request"
// @Success      200  {object}  object{access_token=string,refresh_token=string}
// @Failure      400  {object}  object{error=string}
// @Failure      401  {object}  object{error=string}
// @Failure      500  {object}  object{error=string}
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "refresh_token is required"})
	}

	pair, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired refresh token"})
		}
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("refresh failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(pair)
}

// Logout godoc
// @Summary      Logout (revoke refresh token)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body object{refresh_token=string} true "Logout request"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  object{error=string}
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "refresh_token is required"})
	}

	// ไม่ return error ถ้า token ไม่เจอ (idempotent logout)
	_ = h.authService.Logout(req.RefreshToken)

	log.Info().Str("correlation_id", correlationID).Msg("user logged out")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "logged out successfully"})
}

// LogoutAll godoc
// @Summary      Logout from all devices
// @Tags         auth
// @Produce      json
// @Param        X-User-Id header string true "User ID (injected by gateway)"
// @Success      200  {object}  object{message=string}
// @Failure      401  {object}  object{error=string}
// @Failure      500  {object}  object{error=string}
// @Router       /api/v1/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)
	userID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.authService.LogoutAll(userID); err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("logout all failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	log.Info().Str("correlation_id", correlationID).Uint("user_id", userID).Msg("logged out all devices")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "logged out from all devices"})
}

// GetProfile godoc
// @Summary      Get current user profile
// @Tags         auth
// @Produce      json
// @Param        X-User-Id header string true "User ID (injected by gateway)"
// @Success      200  {object}  object{id=int,username=string,email=string,role=string,created_at=string}
// @Failure      401  {object}  object{error=string}
// @Failure      404  {object}  object{error=string}
// @Router       /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)
	userID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	user, err := h.authService.GetProfile(userID)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("get profile failed")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"group":      user.Group.Name,
		"roles":      user.Group.RoleNames(),
		"created_at": user.CreatedAt,
	})
}

// ChangePassword godoc
// @Summary      Change password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        X-User-Id header string true "User ID (injected by gateway)"
// @Param        request body object{current_password=string,new_password=string} true "Change password request"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  object{error=string}
// @Failure      401  {object}  object{error=string}
// @Failure      500  {object}  object{error=string}
// @Router       /api/v1/auth/password [put]
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)
	userID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "current_password and new_password are required"})
	}
	if len(req.NewPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "new password must be at least 8 characters"})
	}

	if err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrIncorrectPassword) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "current password is incorrect"})
		}
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("change password failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	log.Info().Str("correlation_id", correlationID).Uint("user_id", userID).Msg("password changed")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "password changed — please login again"})
}

// ForgotPassword godoc
// @Summary      Request password reset
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body object{email=string} true "Forgot password request"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  object{error=string}
// @Router       /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}

	token, err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			// ไม่เปิดเผยว่า email ไม่มีในระบบ เพื่อความปลอดภัย
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "if the email exists, a reset link has been sent",
			})
		}
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("forgot password failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	// ในระบบจริงจะส่ง email — ตอนนี้ log token สำหรับ development
	log.Info().
		Str("correlation_id", correlationID).
		Str("reset_token", token).
		Str("email", req.Email).
		Msg("password reset token generated (dev mode — would send email in production)")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "if the email exists, a reset link has been sent",
		"reset_token": token, // TODO: ลบออกเมื่อมี email service จริง
	})
}

// ResetPassword godoc
// @Summary      Reset password with token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body object{token=string,new_password=string} true "Reset password request"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  object{error=string}
// @Failure      401  {object}  object{error=string}
// @Failure      500  {object}  object{error=string}
// @Router       /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Token == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token and new_password are required"})
	}
	if len(req.NewPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "new password must be at least 8 characters"})
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired reset token"})
		}
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("reset password failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	log.Info().Str("correlation_id", correlationID).Msg("password reset successfully")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "password has been reset — please login with your new password"})
}

// ListGroups godoc
// @Summary List groups
// @Description Get all user groups with role mappings
// @Tags auth
// @Produce json
// @Success      200  {object}  object{groups=[]object{id=int,name=string,description=string,is_active=bool,roles=[]object{id=int,name=string},users=[]object{id=int,username=string,email=string}}}
// @Failure      500  {object}  object{error=string}
// @Router /api/v1/auth/groups [get]
func (h *AuthHandler) ListGroups(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)

	groups, err := h.authService.ListGroups()
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("list groups failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	type roleResp struct {
		ID   uint   `json:"ID"`
		Name string `json:"name"`
	}
	type userResp struct {
		ID       uint   `json:"ID"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	type groupResp struct {
		ID          uint       `json:"ID"`
		Name        string     `json:"name"`
		Description string     `json:"description"`
		IsActive    bool       `json:"is_active"`
		Roles       []roleResp `json:"roles"`
		Users       []userResp `json:"users"`
	}

	result := make([]groupResp, len(groups))
	for i, g := range groups {
		roles := make([]roleResp, len(g.Roles))
		for j, r := range g.Roles {
			roles[j] = roleResp{ID: r.ID, Name: r.Name}
		}
		users := make([]userResp, len(g.Users))
		for j, u := range g.Users {
			users[j] = userResp{ID: u.ID, Username: u.Username, Email: u.Email}
		}
		result[i] = groupResp{ID: g.ID, Name: g.Name, Description: g.Description, IsActive: g.IsActive, Roles: roles, Users: users}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"groups": result})
}

// ListRoles godoc
// @Summary List roles
// @Description Get all available roles
// @Tags auth
// @Produce json
// @Success      200  {object}  object{roles=[]object{id=int,name=string}}
// @Failure      500  {object}  object{error=string}
// @Router /api/v1/auth/roles [get]
func (h *AuthHandler) ListRoles(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)

	roles, err := h.authService.ListRoles()
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("list roles failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	type roleResp struct {
		ID   uint   `json:"ID"`
		Name string `json:"name"`
	}
	result := make([]roleResp, len(roles))
	for i, r := range roles {
		result[i] = roleResp{ID: r.ID, Name: r.Name}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"roles": result})
}

// CreateGroup godoc
// @Summary Create group
// @Description Create a new user group with roles and optional user assignments
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object{name=string,description=string,is_active=bool,role_ids=[]int,user_ids=[]int} true "Group data"
// @Success      201  {object}  object{group=object{id=int,name=string,description=string,is_active=bool,roles=[]object{id=int,name=string},users=[]object{id=int,username=string,email=string}}}
// @Failure      400  {object}  object{error=string}
// @Failure      409  {object}  object{error=string}
// @Failure      500  {object}  object{error=string}
// @Router /api/v1/auth/groups [post]
func (h *AuthHandler) CreateGroup(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    *bool  `json:"is_active"`
		RoleIDs     []uint `json:"role_ids"`
		Users       []struct {
			ID uint `json:"id"`
		} `json:"users"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	userIDs := make([]uint, len(req.Users))
	for i, u := range req.Users {
		userIDs[i] = u.ID
	}

	group, err := h.authService.CreateGroup(req.Name, req.Description, isActive, req.RoleIDs, userIDs)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("create group failed")
		if err == service.ErrGroupExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	type roleResp struct {
		ID   uint   `json:"ID"`
		Name string `json:"name"`
	}
	type userResp struct {
		ID       uint   `json:"ID"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	type groupResp struct {
		ID          uint       `json:"ID"`
		Name        string     `json:"name"`
		Description string     `json:"description"`
		IsActive    bool       `json:"is_active"`
		Roles       []roleResp `json:"roles"`
		Users       []userResp `json:"users"`
	}

	roles := make([]roleResp, len(group.Roles))
	for i, r := range group.Roles {
		roles[i] = roleResp{ID: r.ID, Name: r.Name}
	}
	users := make([]userResp, len(group.Users))
	for i, u := range group.Users {
		users[i] = userResp{ID: u.ID, Username: u.Username, Email: u.Email}
	}

	log.Info().Str("correlation_id", correlationID).Uint("group_id", group.ID).Msg("group created")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"group": groupResp{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			IsActive:    group.IsActive,
			Roles:       roles,
			Users:       users,
		},
	})
}

// UpdateGroup godoc
// @Summary Update group
// @Description Update an existing user group (name, description, active status, roles, user assignments)
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object{id=int,name=string,description=string,is_active=bool,role_ids=[]int,user_ids=[]int} true "Group data"
// @Success      200  {object}  object{group=object{id=int,name=string,description=string,is_active=bool,roles=[]object{id=int,name=string},users=[]object{id=int,username=string,email=string}}}
// @Failure      400  {object}  object{error=string}
// @Failure      404  {object}  object{error=string}
// @Failure      500  {object}  object{error=string}
// @Router /api/v1/auth/group [put]
func (h *AuthHandler) UpdateGroup(c *fiber.Ctx) error {
	correlationID := c.Locals("correlationID").(string)

	var req struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    *bool  `json:"is_active"`
		RoleIDs     []uint `json:"role_ids"`
		Users       []struct {
			ID uint `json:"id"`
		} `json:"users"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.ID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "group ID is required"})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	userIDs := make([]uint, len(req.Users))
	for i, u := range req.Users {
		userIDs[i] = u.ID
	}

	group, err := h.authService.UpdateGroup(req.ID, req.Name, req.Description, isActive, req.RoleIDs, userIDs)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("update group failed")
		if err == service.ErrGroupNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	type roleResp struct {
		ID   uint   `json:"ID"`
		Name string `json:"name"`
	}
	type userResp struct {
		ID       uint   `json:"ID"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	type groupResp struct {
		ID          uint       `json:"ID"`
		Name        string     `json:"name"`
		Description string     `json:"description"`
		IsActive    bool       `json:"is_active"`
		Roles       []roleResp `json:"roles"`
		Users       []userResp `json:"users"`
	}
	roles := make([]roleResp, len(group.Roles))
	for i, r := range group.Roles {
		roles[i] = roleResp{ID: r.ID, Name: r.Name}
	}
	users := make([]userResp, len(group.Users))
	for i, u := range group.Users {
		users[i] = userResp{ID: u.ID, Username: u.Username, Email: u.Email}
	}

	log.Info().Str("correlation_id", correlationID).Uint("group_id", group.ID).Msg("group updated")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"group": groupResp{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			IsActive:    group.IsActive,
			Roles:       roles,
			Users:       users,
		},
	})
}

// ── Helper ───────────────────────────────────────────────────────

func extractUserID(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Get("X-User-Id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
