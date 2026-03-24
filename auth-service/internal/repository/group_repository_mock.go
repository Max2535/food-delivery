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
