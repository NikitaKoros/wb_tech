package repository

import (
	"context"
	"database/sql"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
)

type RepositoryProvider interface {
	UpsertOrder(context.Context, *model.Order) (*model.Order, error)
	GetOrderByUID(context.Context, string) (*model.Order, error)
	GetAllOrders(context.Context, int) ([]*model.Order, error)
	GetItemsByOrderUID(context.Context, string, int, int) ([]*model.Item, error)
}

type Querier interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}