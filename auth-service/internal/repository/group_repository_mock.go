package repository

import (
	"auth-service/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) FindByName(name string) (*model.Group, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Group), args.Error(1)
}

func (m *MockGroupRepository) FindByID(id uint) (*model.Group, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Group), args.Error(1)
}

func (m *MockGroupRepository) ListAll() ([]*model.Group, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Group), args.Error(1)
}

func (m *MockGroupRepository) ListRoles() ([]*model.Role, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Role), args.Error(1)
}

func (m *MockGroupRepository) Create(group *model.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockGroupRepository) Update(group *model.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockGroupRepository) Delete(group *model.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockGroupRepository) FindRolesByIDs(ids []uint) ([]model.Role, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Role), args.Error(1)
}
