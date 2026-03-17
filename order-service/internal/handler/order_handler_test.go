package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"order-service/internal/handler"
	"order-service/internal/model"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Service ---

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(order *model.Order, correlationID string) error {
	args := m.Called(order, correlationID)
	return args.Error(0)
}

func (m *MockOrderService) CreateOrderFromRequest(req *model.CreateOrderRequest, correlationID string) (*model.Order, error) {
	args := m.Called(req, correlationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderService) GetAllOrders() ([]model.Order, error) {
	args := m.Called()
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(id uint) (*model.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

// --- Helper ---

func setupApp(svc *MockOrderService) *fiber.App {
	app := fiber.New()
	h := handler.NewOrderHandler(svc)
	api := app.Group("/api/v1/orders")
	api.Post("/", h.CreateOrder)
	api.Get("/", h.GetAllOrders)
	api.Get("/:id", h.GetOrderByID)
	return app
}

// --- Tests ---

func TestCreateOrder_Success(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	resultOrder := &model.Order{
		CustomerID:  "CUST001",
		TotalAmount: 220.0,
		Status:      "pending",
		Items: []model.OrderItem{
			{MenuItemID: 1, MenuItemName: "ผัดกะเพรา", UnitPrice: 50, Quantity: 2, TotalPrice: 100},
			{MenuItemID: 2, MenuItemName: "ต้มยำกุ้ง", UnitPrice: 120, Quantity: 1, TotalPrice: 120},
		},
	}

	mockSvc.On("CreateOrderFromRequest", mock.AnythingOfType("*model.CreateOrderRequest"), mock.AnythingOfType("string")).Return(resultOrder, nil)

	body, _ := json.Marshal(map[string]interface{}{
		"customer_id": "CUST001",
		"items": []map[string]interface{}{
			{"menu_item_id": 1, "menu_item_name": "ผัดกะเพรา", "unit_price": 50, "quantity": 2},
			{"menu_item_id": 2, "menu_item_name": "ต้มยำกุ้ง", "unit_price": 120, "quantity": 1},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestCreateOrder_InvalidJSON(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrder_MissingCustomerID(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	body, _ := json.Marshal(map[string]interface{}{
		"items": []map[string]interface{}{
			{"menu_item_id": 1, "menu_item_name": "ผัดกะเพรา", "unit_price": 50, "quantity": 1},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrder_NoItems(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	body, _ := json.Marshal(map[string]interface{}{
		"customer_id": "CUST001",
		"items":       []map[string]interface{}{},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrder_ServiceError(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	mockSvc.On("CreateOrderFromRequest", mock.AnythingOfType("*model.CreateOrderRequest"), mock.AnythingOfType("string")).Return(nil, errors.New("db error"))

	body, _ := json.Marshal(map[string]interface{}{
		"customer_id": "CUST001",
		"items": []map[string]interface{}{
			{"menu_item_id": 1, "menu_item_name": "ผัดกะเพรา", "unit_price": 50, "quantity": 1},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetAllOrders_ReturnsOrders(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	orders := []model.Order{
		{
			CustomerID:  "CUST001",
			TotalAmount: 100,
			Status:      "pending",
			Items: []model.OrderItem{
				{MenuItemID: 1, MenuItemName: "ผัดกะเพรา", UnitPrice: 50, Quantity: 2, TotalPrice: 100},
			},
		},
	}
	mockSvc.On("GetAllOrders").Return(orders, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string][]model.Order
	json.Unmarshal(respBody, &result)
	assert.Len(t, result["orders"], 1)
	mockSvc.AssertExpectations(t)
}

func TestGetOrderByID_Found(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	order := &model.Order{
		CustomerID:  "CUST001",
		TotalAmount: 100,
		Items: []model.OrderItem{
			{MenuItemID: 1, MenuItemName: "ผัดกะเพรา", UnitPrice: 50, Quantity: 2, TotalPrice: 100},
		},
	}
	mockSvc.On("GetOrderByID", uint(1)).Return(order, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestGetOrderByID_NotFound(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	mockSvc.On("GetOrderByID", uint(99)).Return(nil, errors.New("record not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/99", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestGetOrderByID_InvalidID(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/abc", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
