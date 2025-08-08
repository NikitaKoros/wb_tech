package repository

import (
	"context"
	"database/sql"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
)

type RepositoryProvider interface {
	CreateOrder(context.Context, *model.Order) (*model.Order, error)
	GetOrderByID(context.Context, string) (*model.Order, error)
	GetItemsByOrderUID(context.Context, string, int) ([]*model.Item, error)
}

type Querier interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}