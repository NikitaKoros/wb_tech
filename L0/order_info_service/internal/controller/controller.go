package controller

import (
	"context"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/cache"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/logger"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/repository"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
)

type ControllerProvider interface {
	GetOrderByID(context.Context, string) (*model.Order, error)
	GetItemsByOrderUID(context.Context, string, int) ([]*model.Item, error)
}

type Controller struct {
	repo   repository.RepositoryProvider
	cache  cache.Cache
	logger logger.Logger
}

func NewController(r repository.RepositoryProvider, c cache.Cache, l logger.Logger) *Controller {
	return &Controller{
		repo: r, 
		cache: c,
		logger: l,
	}
}

func (ctrl *Controller) GetOrderByID(ctx context.Context, orderID string) (*model.Order, error) {
	order, err := ctrl.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (ctrl *Controller) GetItemsByOrderUID(ctx context.Context, orderID string, limit int) ([]*model.Item, error) {
	items, err := ctrl.repo.GetItemsByOrderUID(ctx, orderID, limit)
	if err != nil {
		return nil, err
	}
	return items, nil
}
