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

	mockSvc.On("CreateOrder", mock.AnythingOfType("*model.Order"), mock.AnythingOfType("string")).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{
		"customer_id":  "CUST001",
		"total_amount": 150.0,
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

func TestCreateOrder_ServiceError(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	mockSvc.On("CreateOrder", mock.AnythingOfType("*model.Order"), mock.AnythingOfType("string")).Return(errors.New("db error"))

	body, _ := json.Marshal(map[string]interface{}{"customer_id": "CUST002"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetAllOrders_ReturnsOrders(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	orders := []model.Order{{CustomerID: "CUST001", TotalAmount: 100, Status: "pending"}}
	mockSvc.On("GetAllOrders").Return(orders, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result []model.Order
	json.Unmarshal(body, &result)
	assert.Len(t, result, 1)
	mockSvc.AssertExpectations(t)
}

func TestGetOrderByID_Found(t *testing.T) {
	mockSvc := new(MockOrderService)
	app := setupApp(mockSvc)

	order := &model.Order{CustomerID: "CUST001", TotalAmount: 100}
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
