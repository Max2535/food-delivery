package repository

import (
	"auth-service/internal/model"

	"github.com/stretchr/testify/mock"
)

type MockEmailVerificationTokenRepository struct {
	mock.Mock
}

func (m *MockEmailVerificationTokenRepository) Create(token *model.EmailVerificationToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockEmailVerificationTokenRepository) FindByTokenHash(hash string) (*model.EmailVerificationToken, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.EmailVerificationToken), args.Error(1)
}

func (m *MockEmailVerificationTokenRepository) DeleteByTokenHash(hash string) error {
	args := m.Called(hash)
	return args.Error(0)
}

func (m *MockEmailVerificationTokenRepository) DeleteByUserID(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockEmailVerificationTokenRepository) DeleteExpired() error {
	args := m.Called()
	return args.Error(0)
}
