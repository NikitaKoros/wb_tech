package controller

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/cache"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/logger"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/repository"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
)

type ControllerProvider interface {
	GetOrderByUID(context.Context, string) (*model.Order, error)
	GetItemsByOrderUID(context.Context, string, int, int) ([]*model.Item, error)
}

type Controller struct {
	repo   repository.RepositoryProvider
	cache  cache.Cache
	logger logger.Logger
}

func NewController(r repository.RepositoryProvider, c cache.Cache, l logger.Logger) *Controller {
	return &Controller{
		repo:   r,
		cache:  c,
		logger: l,
	}
}

func (ctrl *Controller) GetOrderByUID(ctx context.Context, orderID string) (*model.Order, error) {
	ctrl.logger.Info("controller: request to get order by id",
		zap.String("order_uid", orderID))
	
	order, err := ctrl.cache.GetOrderByUID(orderID)
	if err == nil {
		return order, nil
	}

	order, err = ctrl.repo.GetOrderByUID(ctx, orderID)
	if err != nil {
		logError(ctrl.logger, "controller: failed to get order by id", orderID, err)
		return nil, err
	}
	ctrl.cache.SetOrder(order)
	return order, nil
}

func (ctrl *Controller) GetItemsByOrderUID(ctx context.Context, orderID string, lastID, limit int) ([]*model.Item, error) {
	ctrl.logger.Info("controller: request to get items by order id", 
		zap.String("order_uid", orderID),
		zap.Int("limit", limit))
	
	items, err := ctrl.cache.GetItemsByOrderUID(orderID, lastID, limit)
	if err == nil && len(items) != 0 {
		return items, nil
	}

	items, err = ctrl.repo.GetItemsByOrderUID(ctx, orderID, lastID, limit)
	if err != nil {
		logError(ctrl.logger, "controller: failed to get items", orderID, err)
		return nil, err
	}

	order, err := ctrl.GetOrderByUID(ctx, orderID)
	if err != nil {
		ctrl.logger.Warn("controller: failed to update cache", 
        zap.String("order_uid", orderID), 
        zap.Error(err))
		return items, nil
	}

	order.Items = items
	ctrl.cache.SetOrder(order)
	return items, nil
}

func WarmUpCache(ctx context.Context, repo repository.RepositoryProvider, cache cache.Cache, limit int) (int, error){
	orders, err := repo.GetAllOrders(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to get orders to warmup cache: %w", err)
	}
	
	for _, order := range orders {
		cache.SetOrder(order)
	}
	
	return len(orders), nil
}

func logError(logger logger.Logger, msg string, orderID string, err error) {
	if errors.Is(err, srvcerrors.ErrNotFound) {
		logger.Warn(msg, zap.String("order_uid", orderID), zap.Error(err))
	} else {
		logger.Error(msg, zap.String("order_uid", orderID), zap.Error(err))
	}
}
