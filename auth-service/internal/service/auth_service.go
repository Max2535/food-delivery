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
)

type AuthService interface {
	Register(username, password, email string) (*model.User, error)
	Login(username, password string) (*model.TokenPair, error)
	Refresh(refreshToken string) (*model.TokenPair, error)
	Logout(refreshToken string) error
	LogoutAll(userID uint) error
	GetProfile(userID uint) (*model.User, error)
	ChangePassword(userID uint, currentPassword, newPassword string) error
}

type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.RefreshTokenRepository
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.RefreshTokenRepository) AuthService {
	return &authService{userRepo: userRepo, tokenRepo: tokenRepo}
}

func (s *authService) Register(username, password, email string) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username: username,
		Password: string(hashed),
		Email:    email,
		Role:     model.RoleUser,
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
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"roles":    []string{user.Role},
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
