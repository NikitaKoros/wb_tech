package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/handler"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/logger"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockController struct {
	mock.Mock
}

func (m *MockController) GetOrderByUID(ctx context.Context, orderID string) (*model.Order, error) {
	args := m.Called(ctx, orderID)
	if order, ok := args.Get(0).(*model.Order); ok || args.Get(0) == nil {
		return order, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockController) GetItemsByOrderUID(ctx context.Context, orderID string, lastID, limit int) ([]*model.Item, error) {
	args := m.Called(ctx, orderID, lastID, limit)
	if items, ok := args.Get(0).([]*model.Item); ok || args.Get(0) == nil {
		return items, args.Error(1)
	}

	return nil, args.Error(1)
}

type MockLogger struct{}

func (l *MockLogger) Info(msg string, fields ...logger.Field)  {}
func (l *MockLogger) Error(msg string, fields ...logger.Field) {}
func (l *MockLogger) Debug(msg string, fields ...logger.Field) {}
func (l *MockLogger) Warn(msg string, fields ...logger.Field)  {}

func generateTestOrder(uid string) *model.Order {
	return &model.Order{
		OrderUID:    uid,
		TrackNumber: "TRK-" + uid,
	}
}

func TestHandler_GetOrder_Success(t *testing.T) {
	mockCtrl := new(MockController)

	order := generateTestOrder("ORDER-001")
	mockCtrl.On("GetOrderByUID", mock.Anything, "ORDER-001").Return(order, nil)

	h := handler.NewHandler(mockCtrl, &MockLogger{})

	req := httptest.NewRequest(http.MethodGet, "/api/orders/ORDER-001", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"order_uid":"ORDER-001"`)

	mockCtrl.AssertExpectations(t)
}

func TestHandler_GetOrder_InvalidID(t *testing.T) {
	mockCtrl := new(MockController)
	h := handler.NewHandler(mockCtrl, &MockLogger{})

	req := httptest.NewRequest(http.MethodGet, "/api/orders/%20", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":400`)
	assert.Contains(t, rec.Body.String(), `"message":"Invalid request parameters"`)

	mockCtrl.AssertNotCalled(t, "GetOrderByUID")
}

func TestHandler_GetOrderItems_Success(t *testing.T) {
	mockCtrl := new(MockController)

	items := []*model.Item{
		{ChrtID: 1},
		{ChrtID: 2},
	}
	mockCtrl.On("GetItemsByOrderUID", mock.Anything, "ORDER-001", 0, 10).Return(items, nil)

	h := handler.NewHandler(mockCtrl, &MockLogger{})

	req := httptest.NewRequest(http.MethodGet, "/api/orders/ORDER-001/items?limit=10", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"chrt_id":1`)

	mockCtrl.AssertExpectations(t)
}

func TestHandler_GetOrderItems_NotFound(t *testing.T) {
	mockCtrl := new(MockController)

	mockCtrl.On("GetItemsByOrderUID", mock.Anything, "ORDER-001", 0, 10).Return(nil, srvcerrors.ErrNotFound)

	h := handler.NewHandler(mockCtrl, &MockLogger{})

	req := httptest.NewRequest(http.MethodGet, "/api/orders/ORDER-001/items", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":404`)
	assert.Contains(t, rec.Body.String(), `"message":"Order not found"`)

	mockCtrl.AssertExpectations(t)
}
