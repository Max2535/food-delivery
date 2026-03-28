package service

import (
	"auth-service/internal/model"
	"auth-service/internal/repository"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("username or email already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrIncorrectPassword  = errors.New("current password is incorrect")
	ErrUserNotFound       = errors.New("user not found")
	ErrGroupExists        = errors.New("group name already exists")
	ErrGroupNotFound      = errors.New("group not found")
)

type AuthService interface {
	Register(username, password, email string) (*model.User, error)
	Login(username, password string) (*model.TokenPair, error)
	Refresh(refreshToken string) (*model.TokenPair, error)
	Logout(refreshToken string) error
	LogoutAll(userID uint) error
	GetProfile(userID uint) (*model.User, error)
	ChangePassword(userID uint, currentPassword, newPassword string) error
	ForgotPassword(email string) (string, error)
	ResetPassword(token, newPassword string) error
	ListGroups() ([]*model.Group, error)
	ListRoles() ([]*model.Role, error)
	CreateGroup(name, description string, isActive bool, roleIDs []uint, userIDs []uint) (*model.Group, error)
	UpdateGroup(id uint, name, description string, isActive bool, roleIDs []uint, userIDs []uint) (*model.Group, error)
	DeleteGroup(id uint) error
	GetMenuConfig(userID uint) ([]model.NavGroupResponse, error)
}

type authService struct {
	userRepo       repository.UserRepository
	tokenRepo      repository.RefreshTokenRepository
	resetTokenRepo repository.PasswordResetTokenRepository
	groupRepo      repository.GroupRepository
	navMenuRepo    repository.NavMenuRepository
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.RefreshTokenRepository, opts ...any) AuthService {
	s := &authService{userRepo: userRepo, tokenRepo: tokenRepo}
	for _, opt := range opts {
		switch v := opt.(type) {
		case repository.PasswordResetTokenRepository:
			s.resetTokenRepo = v
		case repository.GroupRepository:
			s.groupRepo = v
		case repository.NavMenuRepository:
			s.navMenuRepo = v
		}
	}
	return s
}

func (s *authService) Register(username, password, email string) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	group, err := s.groupRepo.FindByName(model.GroupUser)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: string(hashed),
		Email:    email,
		GroupID:  group.ID,
		Group:    *group,
	}
	if err := s.userRepo.Create(user); err != nil {
		if isDuplicateError(err) {
			return nil, ErrUserExists
		}
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(username, password string) (*model.TokenPair, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issuePair(user)
}

func (s *authService) Refresh(refreshToken string) (*model.TokenPair, error) {
	hash := hashToken(refreshToken)
	rt, err := s.tokenRepo.FindByTokenHash(hash)
	if err != nil || rt.IsExpired() {
		return nil, ErrInvalidToken
	}
	_ = s.tokenRepo.DeleteByTokenHash(hash)
	return s.issuePair(&rt.User)
}

func (s *authService) Logout(refreshToken string) error {
	return s.tokenRepo.DeleteByTokenHash(hashToken(refreshToken))
}

func (s *authService) LogoutAll(userID uint) error {
	return s.tokenRepo.DeleteByUserID(userID)
}

func (s *authService) GetProfile(userID uint) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *authService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return ErrIncorrectPassword
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(userID, string(hashed))
}

func (s *authService) ForgotPassword(email string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", ErrUserNotFound
	}

	// ลบ token เก่าของ user คนนี้
	_ = s.resetTokenRepo.DeleteByUserID(user.ID)

	rawToken, err := generateRandomToken()
	if err != nil {
		return "", err
	}
	rt := &model.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.resetTokenRepo.Create(rt); err != nil {
		return "", err
	}
	return rawToken, nil
}

func (s *authService) ResetPassword(token, newPassword string) error {
	hash := hashToken(token)
	rt, err := s.resetTokenRepo.FindByTokenHash(hash)
	if err != nil || rt.IsExpired() {
		return ErrInvalidToken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.userRepo.UpdatePassword(rt.UserID, string(hashed)); err != nil {
		return err
	}

	// ลบ token หลังใช้งาน
	_ = s.resetTokenRepo.DeleteByTokenHash(hash)
	// ลบ refresh tokens ทั้งหมดเพื่อบังคับ login ใหม่
	_ = s.tokenRepo.DeleteByUserID(rt.UserID)

	return nil
}

func (s *authService) ListGroups() ([]*model.Group, error) {
	return s.groupRepo.ListAll()
}

func (s *authService) ListRoles() ([]*model.Role, error) {
	return s.groupRepo.ListRoles()
}

func (s *authService) CreateGroup(name, description string, isActive bool, roleIDs []uint, userIDs []uint) (*model.Group, error) {
	roles, err := s.groupRepo.FindRolesByIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	group := &model.Group{
		Name:        name,
		Description: description,
		IsActive:    isActive,
		Roles:       roles,
	}

	if err := s.groupRepo.Create(group); err != nil {
		if isDuplicateError(err) {
			return nil, ErrGroupExists
		}
		return nil, err
	}

	// Assign existing users to this group
	if len(userIDs) > 0 {
		if err := s.userRepo.UpdateGroupID(userIDs, group.ID); err != nil {
			return nil, err
		}
		users, err := s.userRepo.FindByIDs(userIDs)
		if err != nil {
			return nil, err
		}
		group.Users = users
	}

	return group, nil
}

func (s *authService) UpdateGroup(id uint, name, description string, isActive bool, roleIDs []uint, userIDs []uint) (*model.Group, error) {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	roles, err := s.groupRepo.FindRolesByIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	group.Name = name
	group.Description = description
	group.IsActive = isActive
	group.Roles = roles

	if err := s.groupRepo.Update(group); err != nil {
		if isDuplicateError(err) {
			return nil, ErrGroupExists
		}
		return nil, err
	}

	if len(userIDs) > 0 {
		if err := s.userRepo.UpdateGroupID(userIDs, group.ID); err != nil {
			return nil, err
		}
		users, err := s.userRepo.FindByIDs(userIDs)
		if err != nil {
			return nil, err
		}
		group.Users = users
	}

	return group, nil
}

func (s *authService) DeleteGroup(id uint) error {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return ErrGroupNotFound
	}
	return s.groupRepo.Delete(group)
}

func (s *authService) GetMenuConfig(userID uint) ([]model.NavGroupResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	groups, err := s.navMenuRepo.ListAll()
	if err != nil {
		return nil, err
	}

	allResponses := make([]model.NavGroupResponse, len(groups))
	for i, g := range groups {
		allResponses[i] = g.ToResponse()
	}

	roles := user.Group.RoleNames()
	return model.FilterNavMenuByRoles(allResponses, roles), nil
}

// ── Helpers ──────────────────────────────────────────────────────

func (s *authService) issuePair(user *model.User) (*model.TokenPair, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}
	rawRefresh, err := generateRandomToken()
	if err != nil {
		return nil, err
	}
	rt := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.tokenRepo.Create(rt); err != nil {
		return nil, err
	}
	return &model.TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

func (s *authService) generateAccessToken(user *model.User) (string, error) {
	keyPath := os.Getenv("PRIVATE_KEY_PATH")
	if keyPath == "" {
		keyPath = "private_key.pem"
	}
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return "", err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return "", err
	}
	roleNames := make([]string, len(user.Group.Roles))
	for i, r := range user.Group.Roles {
		roleNames[i] = r.Name
	}
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"group":    user.Group.Name,
		"roles":    roleNames,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "default"
	return token.SignedString(privateKey)
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "SQLSTATE 23505")
}
