package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"kitchen-service/internal/handler"
	"kitchen-service/internal/model"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Service ---

type MockKitchenService struct {
	mock.Mock
}

func (m *MockKitchenService) CreateTicket(ticket *model.KitchenTicket) error {
	args := m.Called(ticket)
	return args.Error(0)
}

func (m *MockKitchenService) UpdateStatus(orderID uint, status string) error {
	args := m.Called(orderID, status)
	return args.Error(0)
}

// --- Helper ---

func setupApp(svc *MockKitchenService) *fiber.App {
	app := fiber.New()
	h := handler.NewKitchenHandler(svc)
	api := app.Group("/api/v1/kitchen")
	api.Post("/tickets", h.CreateTicket)
	api.Patch("/tickets/:orderId", h.UpdateStatus)
	return app
}

// --- Tests ---

func TestCreateTicket_Success(t *testing.T) {
	mockSvc := new(MockKitchenService)
	app := setupApp(mockSvc)

	mockSvc.On("CreateTicket", mock.AnythingOfType("*model.KitchenTicket")).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{
		"order_id": 1,
		"items":    `[{"name":"Pad Thai"}]`,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/kitchen/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestCreateTicket_InvalidJSON(t *testing.T) {
	mockSvc := new(MockKitchenService)
	app := setupApp(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/kitchen/tickets", bytes.NewReader([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateTicket_ServiceError(t *testing.T) {
	mockSvc := new(MockKitchenService)
	app := setupApp(mockSvc)

	mockSvc.On("CreateTicket", mock.AnythingOfType("*model.KitchenTicket")).Return(errors.New("db error"))

	body, _ := json.Marshal(map[string]interface{}{"order_id": 2})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/kitchen/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestUpdateStatus_Success(t *testing.T) {
	mockSvc := new(MockKitchenService)
	app := setupApp(mockSvc)

	mockSvc.On("UpdateStatus", uint(1), "Cooking").Return(nil)

	body, _ := json.Marshal(map[string]string{"status": "Cooking"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/kitchen/tickets/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestUpdateStatus_InvalidOrderID(t *testing.T) {
	mockSvc := new(MockKitchenService)
	app := setupApp(mockSvc)

	body, _ := json.Marshal(map[string]string{"status": "Cooking"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/kitchen/tickets/abc", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateStatus_ServiceError(t *testing.T) {
	mockSvc := new(MockKitchenService)
	app := setupApp(mockSvc)

	mockSvc.On("UpdateStatus", uint(99), "Ready").Return(errors.New("not found"))

	body, _ := json.Marshal(map[string]string{"status": "Ready"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/kitchen/tickets/99", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}
