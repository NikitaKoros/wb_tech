package cache

import (
	"testing"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestGetOrderByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cache := NewLocalCache()
		orderID := "order-1"
		order := generateTestOrder(orderID)
		cache.SetOrder(order)

		got, err := cache.GetOrderByID(orderID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.True(t, cmp.Equal(order, got))
	})

	t.Run("not found", func(t *testing.T) {
		cache := NewLocalCache()
		orderID := "nonexistent"
		got, err := cache.GetItemsByOrderUID(orderID)
		require.Error(t, err)
		require.Nil(t, got)
		require.ErrorIs(t, err, srvcerrors.ErrNotFound)
	})
}

func TestGetItemsByOrderUID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cache := NewLocalCache()
		orderID := "order-1"
		order := generateTestOrder(orderID)
		cache.SetOrder(order)
	})

	t.Run("not found", func(t *testing.T) {
		cache := NewLocalCache()
		orderID := "nonexistent"
		items, err := cache.GetItemsByOrderUID(orderID)
		require.Error(t, err)
		require.Nil(t, items)
		require.ErrorIs(t, err, srvcerrors.ErrNotFound)
	})
}

func TestSetOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cache := NewLocalCache()
		orderID := "order-1"
		order := generateTestOrder(orderID)

		cache.SetOrder(order)
		got, err := cache.GetOrderByID(orderID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, orderID, got.OrderUID)
	})
}

func TestClear(t *testing.T) {
	cache := NewLocalCache()
	orderID := "order-1"
	order := generateTestOrder(orderID)

	cache.Clear()

	got, err := cache.GetOrderByID(order.OrderUID)
	require.Error(t, err)
	require.Nil(t, got)
	require.ErrorIs(t, err, srvcerrors.ErrNotFound)
}

func generateTestOrder(uid string) *model.Order {
	return &model.Order{
		OrderUID:    uid,
		TrackNumber: "track-123",
		Items: []*model.Item{
			{
				ChrtID: 1001,
				Name:   "Test Item",
			},
		},
	}
}
