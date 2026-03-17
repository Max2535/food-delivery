package service_test

import (
	"errors"
	"testing"

	"kitchen-service/internal/model"
	"kitchen-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repository ---

type MockKitchenRepository struct {
	mock.Mock
}

func (m *MockKitchenRepository) Create(ticket *model.KitchenTicket) error {
	args := m.Called(ticket)
	return args.Error(0)
}

func (m *MockKitchenRepository) UpdateStatus(orderID uint, status string) error {
	args := m.Called(orderID, status)
	return args.Error(0)
}

func (m *MockKitchenRepository) GetByOrderID(orderID uint) (*model.KitchenTicket, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.KitchenTicket), args.Error(1)
}

func (m *MockKitchenRepository) GetQueue() ([]*model.KitchenTicket, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.KitchenTicket), args.Error(1)
}

// --- Tests ---

func TestCreateTicket_SetsStatusPending(t *testing.T) {
	mockRepo := new(MockKitchenRepository)
	svc := service.NewKitchenService(mockRepo)

	ticket := &model.KitchenTicket{OrderID: 1, Items: `[{"name":"Pad Thai"}]`}
	mockRepo.On("Create", ticket).Return(nil)

	err := svc.CreateTicket(ticket)

	assert.NoError(t, err)
	assert.Equal(t, model.StatusPending, ticket.Status) // Service must set status to "Pending"
	assert.Equal(t, model.PriorityNormal, ticket.Priority) // Must default to Normal
	mockRepo.AssertExpectations(t)
}

func TestCreateTicket_ReturnsRepositoryError(t *testing.T) {
	mockRepo := new(MockKitchenRepository)
	svc := service.NewKitchenService(mockRepo)

	ticket := &model.KitchenTicket{OrderID: 2}
	mockRepo.On("Create", ticket).Return(errors.New("db error"))

	err := svc.CreateTicket(ticket)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestUpdateStatus_Success(t *testing.T) {
	mockRepo := new(MockKitchenRepository)
	svc := service.NewKitchenService(mockRepo)

	mockRepo.On("UpdateStatus", uint(1), "Cooking").Return(nil)

	err := svc.UpdateStatus(1, "Cooking")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateStatus_ReturnsError(t *testing.T) {
	mockRepo := new(MockKitchenRepository)
	svc := service.NewKitchenService(mockRepo)

	mockRepo.On("UpdateStatus", uint(99), "Ready").Return(errors.New("not found"))

	err := svc.UpdateStatus(99, "Ready")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateStatus_Ready_LogsEvent(t *testing.T) {
	mockRepo := new(MockKitchenRepository)
	svc := service.NewKitchenService(mockRepo)

	// When status is "Ready to Serve", service should still succeed and log the event
	mockRepo.On("UpdateStatus", uint(3), model.StatusReadyToServe).Return(nil)

	err := svc.UpdateStatus(3, model.StatusReadyToServe)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
