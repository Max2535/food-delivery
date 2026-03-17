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

func TestCreateOrderFromRequest_Success(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	req := &model.CreateOrderRequest{
		CustomerID: "CUST001",
		Items: []model.CreateOrderItemRequest{
			{MenuItemID: 1, MenuItemName: "ผัดกะเพรา", UnitPrice: 50, Quantity: 2},
			{MenuItemID: 2, MenuItemName: "ต้มยำกุ้ง", UnitPrice: 120, Quantity: 1},
		},
	}

	mockRepo.On("Create", mock.AnythingOfType("*model.Order")).Return(nil)

	order, err := svc.CreateOrderFromRequest(req, "test-correlation-id")

	// Kitchen publish จะ fail (ไม่มี RabbitMQ) แต่ order ยังถูก save ได้
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, "pending", order.Status)
	assert.Equal(t, "CUST001", order.CustomerID)
	assert.Equal(t, 220.0, order.TotalAmount)
	assert.Len(t, order.Items, 2)
	assert.Equal(t, 100.0, order.Items[0].TotalPrice) // 50*2
	assert.Equal(t, 120.0, order.Items[1].TotalPrice) // 120*1
	mockRepo.AssertExpectations(t)
}

func TestCreateOrderFromRequest_EmptyCustomerID(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	req := &model.CreateOrderRequest{
		CustomerID: "",
		Items: []model.CreateOrderItemRequest{
			{MenuItemID: 1, MenuItemName: "ผัดกะเพรา", UnitPrice: 50, Quantity: 1},
		},
	}

	order, err := svc.CreateOrderFromRequest(req, "test-correlation-id")

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "customer_id is required")
}

func TestCreateOrderFromRequest_NoItems(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	req := &model.CreateOrderRequest{
		CustomerID: "CUST001",
		Items:      []model.CreateOrderItemRequest{},
	}

	order, err := svc.CreateOrderFromRequest(req, "test-correlation-id")

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "at least one item is required")
}

func TestCreateOrderFromRequest_RepoError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	req := &model.CreateOrderRequest{
		CustomerID: "CUST001",
		Items: []model.CreateOrderItemRequest{
			{MenuItemID: 1, MenuItemName: "ผัดกะเพรา", UnitPrice: 50, Quantity: 1},
		},
	}

	mockRepo.On("Create", mock.AnythingOfType("*model.Order")).Return(errors.New("db error"))

	order, err := svc.CreateOrderFromRequest(req, "test-correlation-id")

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, "db error", err.Error())
}

func TestCreateOrder_SetsDefaultStatus(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	order := &model.Order{CustomerID: "CUST001", TotalAmount: 100}

	// Expect Create to be called with the order, return nil error
	mockRepo.On("Create", order).Return(nil)

	err := svc.CreateOrder(order, "test-correlation-id")

	if err != nil {
		assert.Contains(t, err.Error(), "Kitchen service is unavailable")
	} else {
		t.Log("Expected Kitchen service unavailable error, but got nil (RabbitMQ is somehow running?)")
	}
	assert.Equal(t, "pending", order.Status) // status must be set to "pending"
	mockRepo.AssertExpectations(t)
}

func TestCreateOrder_PreservesExistingStatus(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	order := &model.Order{CustomerID: "CUST002", TotalAmount: 200, Status: "confirmed"}
	mockRepo.On("Create", order).Return(nil)

	err := svc.CreateOrder(order, "test-correlation-id")

	if err != nil {
		assert.Contains(t, err.Error(), "Kitchen service is unavailable")
	} else {
		t.Log("Expected Kitchen service unavailable error, but got nil (RabbitMQ is somehow running?)")
	}
	assert.Equal(t, "confirmed", order.Status) // original status kept
	mockRepo.AssertExpectations(t)
}

func TestCreateOrder_ReturnsError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	order := &model.Order{CustomerID: "CUST003"}
	mockRepo.On("Create", order).Return(errors.New("db error"))

	err := svc.CreateOrder(order, "test-correlation-id")

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
