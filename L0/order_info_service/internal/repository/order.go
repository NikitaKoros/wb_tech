package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/dto"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
)

type OrderRepository struct {
	db *sql.DB
}

const (
	insertIntoOrdersQuery = `INSERT INTO orders
		(order_uid, track_number, entry, locale, internal_signature,
    		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
      	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
       	RETURNING order_uid, track_number, entry, locale, internal_signature,
      		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard`

	insertIntoDeliveriesQuery = `INSERT INTO deliveries
		(order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING order_uid, name, phone, zip, city, address, region, email`

	insertIntoPaymentsQuery = `INSERT INTO payments
		(transaction, request_id, currency, provider, amount,
    		payment_dt, bank, delivery_cost, goods_total, custom_fee)
    	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    	RETURNING transaction, request_id, currency, provider, amount,
     		payment_dt, bank, delivery_cost, goods_total, custom_fee`

	insertIntoItemsQuery = `INSERT INTO items
		(order_uid, chrt_id, track_number, price, rid,
    		name, sale, size, total_price, nm_id, brand, status)
      	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
       	RETURNING id`

	getOrderByIDQuery = `SELECT * FROM orders
		WHERE order_uid = $1`
		
	getAllOrdersQuery = `SELECT * FROM orders
		ORDER BY order_uid
		LIMIT $1`

	getDeliveryByOrderUIDQuery = `SELECT * FROM deliveries
		WHERE order_uid = $1`

	getPaymentByOrderUIDQuery = `SELECT * FROM payments
		WHERE order_uid = $1`

	getItemsByOrderUIDQuery = `SELECT * FROM items
		WHERE order_uid = $1
		ORDER BY id
		LIMIT $2`
)

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *model.Order) (newOrder *model.Order, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		err = r.finishTransaction(tx, err)
	}()

	row := tx.QueryRowContext(ctx, insertIntoOrdersQuery,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if newOrder, err = dto.ScanOrderFromRow(row); err != nil {
		return nil, fmt.Errorf("failed to insert into orders while creating order: %w", err)
	}

	if _, err := tx.ExecContext(ctx, insertIntoDeliveriesQuery,
		order.Delivery.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	); err != nil {
		return nil, fmt.Errorf("failed to insert into deliveries while creating order: %w", err)
	}
	newOrder.Delivery = order.Delivery
	
	if _, err := tx.ExecContext(ctx, insertIntoPaymentsQuery, 
		order.Payment.Transaction,   
		order.Payment.RequestID,     
		order.Payment.Currency,      
		order.Payment.Provider,      
		order.Payment.Amount,       
		order.Payment.PaymentDT,     
		order.Payment.Bank,          
		order.Payment.DeliveryCost,  
		order.Payment.GoodsTotal,    
		order.Payment.CustomFee,     
	); err != nil {
		return nil, fmt.Errorf("failed to insert into payments while creating order: %w", err)
	}
	newOrder.Payment = order.Payment
	
	for i, item := range order.Items {
		row := tx.QueryRowContext(ctx, insertIntoItemsQuery, 
			item.OrderUID,
			item.ChrtID,      
			item.TrackNumber, 
			item.Price,   
			item.RID,         
			item.Name,        
			item.Sale,     
			item.Size,        
			item.TotalPrice, 
			item.NmID,        
			item.Brand,       
			item.Status,      
		)
		
		if err := row.Scan(&order.Items[i].ID); err != nil {
			return nil, fmt.Errorf("failed to insert into items while creating order: %w", err)
		}
	}
	newOrder.Items = order.Items

	return newOrder, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID string) (order *model.Order, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		err = r.finishTransaction(tx, err)
	}()

	order, err = r.getOrderByOrderUID(ctx, tx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by id %s: %w", orderID, err)
	}

	delivery, err := r.getDeliveryByOrderUID(ctx, tx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery by order_uid %s: %w", orderID, err)
	}

	payment, err := r.getPaymentByOrderUID(ctx, tx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by order id %s: %w", orderID, err)
	}

	order.Delivery = *delivery
	order.Payment = *payment

	return order, nil
}

func (r *OrderRepository) GetAllOrders(ctx context.Context, limit int) (orders []*model.Order, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	defer func() {
		err = r.finishTransaction(tx, err)
	}()
	
	orders = make([]*model.Order, 0)
	orders, err = r.getAllOrders(ctx, tx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}
	
	for i, order := range orders {
		delivery, err := r.getDeliveryByOrderUID(ctx, tx, order.OrderUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get delivery of order %s: %w", order.OrderUID, err)
		}
		
		payment, err := r.getPaymentByOrderUID(ctx, tx, order.OrderUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment of order %s: %w", order.OrderUID, err)
		}
		
		items, err := r.getItemsByOrderUID(ctx, tx, order.OrderUID, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get items for order %s: %w", order.OrderUID, err)
		}
		
		orders[i].Delivery = *delivery
		orders[i].Payment = *payment
		orders[i].Items = items
	}
	
	return orders, nil
}

func (r *OrderRepository) GetItemsByOrderUID(ctx context.Context, orderID string, limit int) (items []*model.Item, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func(){
		err = r.finishTransaction(tx, err)
	}()

	items = make([]*model.Item, 0, limit)
	items, err = r.getItemsByOrderUID(ctx, tx, orderID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get items of order %s: %w", orderID, err)
	}

	return items, nil
}

func (r *OrderRepository) getOrderByOrderUID(ctx context.Context, q Querier, orderID string) (*model.Order, error) {
	row := q.QueryRowContext(ctx, getOrderByIDQuery, orderID)
	order, err := dto.ScanOrderFromRow(row)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *OrderRepository) getAllOrders(ctx context.Context, q Querier, limit int) ([]*model.Order, error) {
	orders := make([]*model.Order, 0)
	
	rows, err := q.QueryContext(ctx, getAllOrdersQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		order, err := dto.ScanOrderFromRow(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return orders, nil
}

func (r *OrderRepository) getDeliveryByOrderUID(ctx context.Context, q Querier, orderID string) (*model.Delivery, error) {
	row := q.QueryRowContext(ctx, getDeliveryByOrderUIDQuery, orderID)
	delivery, err := dto.ScanDeliveryFromRow(row)
	if err != nil {
		return nil, err
	}
	return delivery, nil
}

func (r *OrderRepository) getPaymentByOrderUID(ctx context.Context, q Querier, orderID string) (*model.Payment, error) {
	row := q.QueryRowContext(ctx, getPaymentByOrderUIDQuery, orderID)
	payment, err := dto.ScanPaymentFromRow(row)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (r *OrderRepository) getItemsByOrderUID(ctx context.Context, q Querier, orderID string, limit int) ([]*model.Item, error) {
	var items []*model.Item

	rows, err := q.QueryContext(ctx, getItemsByOrderUIDQuery, orderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item, err := dto.ScanItemFromRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *OrderRepository) finishTransaction(tx *sql.Tx, origErr error) error {
	if origErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%v; rollback failed: %w", origErr, rbErr)
		}
		return fmt.Errorf("%w; rollback succeeded", origErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("commit failed: %w", commitErr)
	}
	return nil
}
