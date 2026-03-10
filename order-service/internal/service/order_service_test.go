package service_test

import (
	"errors"
	"testing"

	"order-service/internal/model"
	"order-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repository ---

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(order *model.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) FindAll() ([]model.Order, error) {
	args := m.Called()
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByID(id uint) (*model.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

// --- Tests ---

func TestCreateOrder_SetsDefaultStatus(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	order := &model.Order{CustomerID: "CUST001", TotalAmount: 100}

	// Expect Create to be called with the order, return nil error
	mockRepo.On("Create", order).Return(nil)

	err := svc.CreateOrder(order)

	assert.NoError(t, err)
	assert.Equal(t, "pending", order.Status) // status must be set to "pending"
	mockRepo.AssertExpectations(t)
}

func TestCreateOrder_PreservesExistingStatus(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	order := &model.Order{CustomerID: "CUST002", TotalAmount: 200, Status: "confirmed"}
	mockRepo.On("Create", order).Return(nil)

	err := svc.CreateOrder(order)

	assert.NoError(t, err)
	assert.Equal(t, "confirmed", order.Status) // original status kept
	mockRepo.AssertExpectations(t)
}

func TestCreateOrder_ReturnsError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	order := &model.Order{CustomerID: "CUST003"}
	mockRepo.On("Create", order).Return(errors.New("db error"))

	err := svc.CreateOrder(order)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestGetAllOrders_ReturnsList(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	expected := []model.Order{
		{CustomerID: "CUST001", TotalAmount: 100, Status: "pending"},
		{CustomerID: "CUST002", TotalAmount: 200, Status: "completed"},
	}
	mockRepo.On("FindAll").Return(expected, nil)

	orders, err := svc.GetAllOrders()

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetOrderByID_ReturnsOrder(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	expected := &model.Order{CustomerID: "CUST001", TotalAmount: 100}
	mockRepo.On("FindByID", uint(1)).Return(expected, nil)

	order, err := svc.GetOrderByID(1)

	assert.NoError(t, err)
	assert.Equal(t, expected.CustomerID, order.CustomerID)
	mockRepo.AssertExpectations(t)
}

func TestGetOrderByID_ReturnsError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	mockRepo.On("FindByID", uint(99)).Return(nil, errors.New("not found"))

	order, err := svc.GetOrderByID(99)

	assert.Error(t, err)
	assert.Nil(t, order)
	mockRepo.AssertExpectations(t)
}
