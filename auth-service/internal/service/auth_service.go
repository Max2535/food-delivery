package service

import (
	"auth-service/internal/model"
	"auth-service/internal/repository"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(req model.LoginRequest) (*model.LoginResponse, error)
	Register(user *model.User) error
}

type authService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Register(user *model.User) error {
	log.Info().Str("username", user.Username).Msg("Registering user")

	// Default role to "user" if not provided
	if user.Role == "" {
		user.Role = model.RoleUser
	}

	// Validate role
	switch user.Role {
	case model.RoleUser, model.RoleAdmin, model.RoleRider, model.RoleMerchant, model.RoleCustomer:
		// Valid role
	default:
		log.Warn().Str("role", user.Role).Msg("Invalid role provided during registration")
		return errors.New("invalid role: " + user.Role)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return err
	}
	user.Password = string(hashedPassword)
	err = s.repo.Create(user)
	if err != nil {
		log.Error().Err(err).Str("username", user.Username).Msg("Failed to create user in DB")
		return err
	}
	log.Info().Str("username", user.Username).Uint("id", user.ID).Msg("User registered successfully")
	return nil
}

func (s *authService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	log.Info().Str("username", req.Username).Msg("Login attempt")
	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		log.Warn().Err(err).Str("username", req.Username).Msg("User not found during login")
		return nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Warn().Str("username", req.Username).Msg("Password mismatch")
		return nil, errors.New("invalid username or password")
	}

	token, err := s.generateToken(user)
	if err != nil {
		log.Error().Err(err).Str("username", req.Username).Msg("Failed to generate token")
		return nil, err
	}

	log.Info().Str("username", req.Username).Msg("Login successful")
	return &model.LoginResponse{Token: token, Role: user.Role}, nil
}

func (s *authService) generateToken(user *model.User) (string, error) {
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

	role := user.Role
	if role == "" {
		role = model.RoleUser
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"roles":    []string{role},
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "default"
	return token.SignedString(privateKey)
}
