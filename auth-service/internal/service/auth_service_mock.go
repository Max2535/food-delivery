package service

import (
	"auth-service/internal/model"

	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(username, password, email string) (*model.User, error) {
	args := m.Called(username, password, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) Login(username, password string) (*model.TokenPair, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TokenPair), args.Error(1)
}

func (m *MockAuthService) Refresh(refreshToken string) (*model.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TokenPair), args.Error(1)
}

func (m *MockAuthService) Logout(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) LogoutAll(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) GetProfile(userID uint) (*model.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	args := m.Called(userID, currentPassword, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) ForgotPassword(email string) (string, error) {
	args := m.Called(email)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ResetPassword(token, newPassword string) error {
	args := m.Called(token, newPassword)
	return args.Error(0)
}
