package repository

import (
	"auth-service/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByIDs(ids []uint) ([]model.User, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(id uint, hashedPassword string) error {
	args := m.Called(id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateGroupID(userIDs []uint, groupID uint) error {
	args := m.Called(userIDs, groupID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateIsVerified(id uint, isVerified bool) error {
	args := m.Called(id, isVerified)
	return args.Error(0)
}
