package controller_test

import (
	"context"
	"errors"
	"testing"

	//"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/cache"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/controller"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/logger"
	//"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/repository"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetOrderByUID(ctx context.Context, orderID string) (*model.Order, error) {
    args := m.Called(ctx, orderID)
    if order, ok := args.Get(0).(*model.Order); ok || args.Get(0) == nil {
        return order, args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *MockRepository) GetItemsByOrderUID(ctx context.Context, orderID string, lastID, limit int) ([]*model.Item, error) {
    args := m.Called(ctx, orderID, lastID, limit)
    if items, ok := args.Get(0).([]*model.Item); ok || args.Get(0) == nil {
        return items, args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *MockRepository) GetAllOrders(ctx context.Context, limit int) ([]*model.Order, error) {
    args := m.Called(ctx, limit)
    if orders, ok := args.Get(0).([]*model.Order); ok || args.Get(0) == nil {
        return orders, args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *MockRepository) UpsertOrder(ctx context.Context, o *model.Order) (*model.Order, error) {
    args := m.Called(ctx, o)
    if order, ok := args.Get(0).(*model.Order); ok || args.Get(0) == nil {
        return order, args.Error(1)
    }
    return nil, args.Error(1)
}

type MockCache struct {
	mock.Mock
}

type MockLogger struct{}

func (n *MockLogger) Info(msg string, fields ...logger.Field) {}
func (n *MockLogger) Error(msg string, fields ...logger.Field) {}
func (n *MockLogger) Debug(msg string, fields ...logger.Field) {}
func (n *MockLogger) Warn(msg string, fields ...logger.Field) {}

func (m *MockCache) GetOrderByUID(orderID string) (*model.Order, error) {
    args := m.Called(orderID)
    if order, ok := args.Get(0).(*model.Order); ok || args.Get(0) == nil {
        return order, args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *MockCache) GetItemsByOrderUID(orderID string) ([]*model.Item, error) {
    args := m.Called(orderID)
    if items, ok := args.Get(0).([]*model.Item); ok || args.Get(0) == nil {
        return items, args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *MockCache) SetOrder(order *model.Order) {
	m.Called(order)
}

func (m *MockCache) Clear() {
	m.Called()
}

func generateTestOrder(uid string, itemCount int) *model.Order {
	order := &model.Order{
		OrderUID:    uid,
		TrackNumber: "TRACK-" + uid,
		Items:       make([]*model.Item, 0, itemCount),
	}

	for i := 0; i < itemCount; i++ {
		order.Items = append(order.Items, &model.Item{
			ChrtID: int(i + 1),
		})
	}

	return order
}

func generateTestItems(count int) []*model.Item {
	items := make([]*model.Item, 0, count)
	for i := 0; i < count; i++ {
		items = append(items, &model.Item{
			ChrtID: int(i + 1),
		})
	}
	return items
}

func TestGetOrderByUID_CacheMiss_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockLogger := new(MockLogger)
	
	order := generateTestOrder("ORDER-001", 2)
	
	mockCache.On("GetOrderByUID", "ORDER-001").Return(nil, errors.New("not found"))
	mockRepo.On("GetOrderByUID", mock.Anything, "ORDER-001").Return(order, nil)
	mockCache.On("SetOrder", order).Return()
	
	ctrl := controller.NewController(mockRepo, mockCache, mockLogger)
	
	ctx := context.Background()
	result, err := ctrl.GetOrderByUID(ctx, "ORDER-001")
	
	require.NoError(t, err)
	assert.Equal(t, order, result)
}

func TestGetOrderByUID_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockLogger := new(MockLogger)
	
	mockCache.On("GetOrderByUID", "ORDER-001").Return(nil, errors.New("not found"))
	mockRepo.On("GetOrderByUID", mock.Anything, "ORDER-001").Return(nil, srvcerrors.ErrDatabase)
	
	ctrl := controller.NewController(mockRepo, mockCache, mockLogger)
	
	ctx := context.Background()
	_, err := ctrl.GetOrderByUID(ctx, "ORDER-001")
	
	require.Error(t, err)
	assert.ErrorIs(t, err, srvcerrors.ErrDatabase)
}

func TestGetItemsByOrderUID_CacheMiss_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockLogger := new(MockLogger)
	
	
	items := generateTestItems(5)
	order := generateTestOrder("ORDER-001", 5)
	
	mockCache.On("GetItemsByOrderUID", "ORDER-001").Return(nil, errors.New("not found"))
	mockCache.On("GetOrderByUID", "ORDER-001").Return(nil, errors.New("not found"))
	mockRepo.On("GetItemsByOrderUID", mock.Anything, "ORDER-001", 0, 10).Return(items, nil)
	mockRepo.On("GetOrderByUID", mock.Anything, "ORDER-001").Return(order, nil)
	mockCache.On("SetOrder", order).Return()
	
	ctrl := controller.NewController(mockRepo, mockCache, mockLogger)
	
	ctx := context.Background()
	result, err := ctrl.GetItemsByOrderUID(ctx, "ORDER-001", 0, 10)
	
	require.NoError(t, err)
	assert.Equal(t, items, result)
	
	assert.Equal(t, items, order.Items)
}

func TestGetItemsByOrderUID_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockLogger := new(MockLogger)
	
	mockCache.On("GetItemsByOrderUID", "ORDER-001").Return(nil, errors.New("not found"))
	mockRepo.On("GetItemsByOrderUID", mock.Anything, "ORDER-001", 0, 10).Return(nil, srvcerrors.ErrDatabase)
	
	ctrl := controller.NewController(mockRepo, mockCache, mockLogger)
	
	ctx := context.Background()
	_, err := ctrl.GetItemsByOrderUID(ctx, "ORDER-001", 0, 10)
	
	require.Error(t, err)
	assert.ErrorIs(t, err, srvcerrors.ErrDatabase)
}

func TestGetItemsByOrderUID_GetOrderError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockLogger := new(MockLogger)
	
	items := generateTestItems(5)
	
	mockCache.On("GetItemsByOrderUID", "ORDER-001").Return(nil, errors.New("not found"))
	mockCache.On("GetOrderByUID", "ORDER-001").Return(nil, errors.New("not found"))
	mockRepo.On("GetItemsByOrderUID", mock.Anything, "ORDER-001", 0, 10).Return(items, nil)
	mockRepo.On("GetOrderByUID", mock.Anything, "ORDER-001").Return(nil, srvcerrors.ErrNotFound)
	
	ctrl := controller.NewController(mockRepo, mockCache, mockLogger)
	
	ctx := context.Background()
	result, err := ctrl.GetItemsByOrderUID(ctx, "ORDER-001", 0, 10)
	
	require.NoError(t, err)
	assert.Equal(t, items, result)
}

func TestWarmUpCache_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	
	orders := []*model.Order{
		generateTestOrder("ORDER-001", 2),
		generateTestOrder("ORDER-002", 3),
	}
	
	mockRepo.On("GetAllOrders", mock.Anything, 10).Return(orders, nil)
	mockCache.On("SetOrder", orders[0]).Return()
	mockCache.On("SetOrder", orders[1]).Return()
	
	ctx := context.Background()
	err := controller.WarmUpCache(ctx, mockRepo, mockCache, 10)
	
	require.NoError(t, err)
}

func TestWarmUpCache_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	
	mockRepo.On("GetAllOrders", mock.Anything, 10).Return(nil, srvcerrors.ErrDatabase)
	
	ctx := context.Background()
	err := controller.WarmUpCache(ctx, mockRepo, mockCache, 10)
	
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get orders to warmup cache")
	assert.ErrorIs(t, err, srvcerrors.ErrDatabase)
}