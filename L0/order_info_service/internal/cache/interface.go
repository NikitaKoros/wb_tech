package cache

import (
	"context"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
)

type Cache interface {
	GetOrderByID(context.Context, string) (*model.Order, error)
	GetItemsByOrderUID(context.Context, string) ([]*model.Item, error)
	SetOrder(context.Context, *model.Order)
	Clear()
}