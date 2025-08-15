package cache

import (
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
)

type Cache interface {
	GetOrderByUID(string) (*model.Order, error)
	GetItemsByOrderUID(string) ([]*model.Item, error)
	SetOrder(*model.Order)
	Clear()
}