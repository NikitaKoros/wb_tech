package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/repository"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
)

type LocalCache struct {
	orders map[string]*model.Order
	mu sync.RWMutex
}

func NewLocalCache() *LocalCache{
	return &LocalCache{
		orders: make(map[string]*model.Order),
	}
}

func WarmUpCache(ctx context.Context, repo repository.RepositoryProvider, cache Cache, limit int) (error){
	orders, err := repo.GetAllOrders(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to get orders to warmup cache: %w", err)
	}
	
	for _, order := range orders {
		cache.SetOrder(order)
	}
	
	return nil
}
 
func (l *LocalCache) GetOrderByID(orderID string) (*model.Order, error){
	l.mu.Lock()
	defer l.mu.Unlock()
	
	order, ok := l.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("%w: order %s not found in cache", srvcerrors.ErrNotFound, orderID)
	}
	return order, nil
}

func (l *LocalCache) GetItemsByOrderUID(orderID string, lastID, limit int) ([]*model.Item, error){
	l.mu.Lock()
	defer l.mu.Unlock()
	
	order, ok := l.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("%w: failed to get items of order %s: order not found in cache", srvcerrors.ErrNotFound, orderID)
	}
	
	if len(order.Items) == 0 {
		return []*model.Item{}, nil
	}
	
	startIndex := 0
	if lastID > 0 {
		found := false
		for i, item := range order.Items {
			if item.ID > lastID {
				startIndex = i
				found = true
				break
			}
		}
		if !found {
			return []*model.Item{}, nil
		}
	}
	
	endIndex := startIndex + limit
	endIndex = min(endIndex, len(order.Items))
	
	return order.Items[startIndex:endIndex], nil
}

func (l *LocalCache) SetOrder(order *model.Order) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.orders[order.OrderUID] = order
}

func (l *LocalCache) Clear(){
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.orders = make(map[string]*model.Order)
}