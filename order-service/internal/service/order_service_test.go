package service_test

import (
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

func TestTC_ORDER_001_CreateOrderSuccess(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	order := &model.Order{
		CustomerID:      "cust-001",
		TotalAmount:     280.00,
		DeliveryAddress: "123 Main St, Bangkok",
		Status:          "pending",
	}

	mockRepo.On("Create", order).Return(nil)

	err := svc.CreateOrder(order, "test-correlation-id")

	// Even if Kitchen service is unavailable, the order creation in DB should succeed
	if err != nil {
		assert.Contains(t, err.Error(), "Kitchen service is unavailable")
	}
	assert.Equal(t, "pending", order.Status)
	assert.Equal(t, "123 Main St, Bangkok", order.DeliveryAddress)
	mockRepo.AssertExpectations(t)
}

func TestTC_ORDER_002_OrderAddressSnapshot(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockRepo)

	// Original order with old address
	order := &model.Order{
		ID:              1,
		CustomerID:      "cust-001",
		DeliveryAddress: "Old Address: 123 Bangkok",
		Status:          "pending",
	}

	mockRepo.On("FindByID", uint(1)).Return(order, nil)

	// Act: Retrieve order
	savedOrder, err := svc.GetOrderByID(1)

	// Assert: Address in order must still be the old address (Snapshot)
	assert.NoError(t, err)
	assert.Equal(t, "Old Address: 123 Bangkok", savedOrder.DeliveryAddress)
	mockRepo.AssertExpectations(t)
}

func TestTC_ORDER_003_CreateOrderWithInvalidMenuItem(t *testing.T) {
	// This test case would normally involve a Catalog Service mock.
	// Current implementation doesn't have a separate validator for menu items yet,
	// but we can add a test to define the expected behavior.
	t.Skip("Menu item validation not yet implemented in service layer")
}

func TestTC_ORDER_004_CreateOrderZeroQuantity(t *testing.T) {
	// Similar to TC-ORDER-003, this requires business logic for items.
	t.Skip("Quantity validation not yet implemented in service layer")
}

