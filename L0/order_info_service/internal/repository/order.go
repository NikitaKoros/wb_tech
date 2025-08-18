package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/dto"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
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

	updateOrderQuery = `UPDATE orders
    	SET track_number = $1, entry = $2, locale = $3, internal_signature = $4,
   			customer_id = $5, delivery_service = $6, shardkey = $7, sm_id = $8, date_created = $9, oof_shard = $10
     	WHERE order_uid = $11
      	RETURNING order_uid, track_number, entry, locale, internal_signature,
     		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard`

	updateDeliveryQuery = `UPDATE deliveries
    	SET name = $1, phone = $2, zip = $3, city = $4, address = $5, region = $6, email = $7
     	WHERE order_uid = $8
      	RETURNING order_uid, name, phone, zip, city, address, region, email`

	updatePaymentQuery = `UPDATE payments
    	SET request_id = $1, currency = $2, provider = $3, amount = $4,
     		payment_dt = $5, bank = $6, delivery_cost = $7, goods_total = $8, custom_fee = $9
     	WHERE transaction = $10
      	RETURNING transaction, request_id, currency, provider, amount,
       		payment_dt, bank, delivery_cost, goods_total, custom_fee`

	updateItemQuery = `UPDATE items
    	SET order_uid = $1, chrt_id = $2, track_number = $3, price = $4, rid = $5,
       		name = $6, sale = $7, size = $8, total_price = $9, nm_id = $10, brand = $11, status = $12
     	WHERE id = $13
      	RETURNING order_uid, chrt_id, track_number, price, rid,
         		name, sale, size, total_price, nm_id, brand, status`

	deleteItemsQuery = `DELETE FROM items WHERE order_uid = $1`

	getOrderByIDQuery = `SELECT * FROM orders
		WHERE order_uid = $1`

	getAllOrdersQuery = `SELECT * FROM orders
		ORDER BY order_uid
		LIMIT $1`

	getDeliveryByOrderUIDQuery = `SELECT * FROM deliveries
		WHERE order_uid = $1`

	getPaymentByOrderUIDQuery = `SELECT * FROM payments
		WHERE transaction = $1`

	getAllItemsByOrderUIDQuery = `SELECT * FROM items
		WHERE order_uid = $1
		ORDER BY id
		LIMIT $2`

	getItemsByOrderUIDQuery = `SELECT * FROM items
		WHERE order_uid = $1 AND id > $2
		ORDER BY id
		LIMIT $3`
)

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) UpsertOrder(ctx context.Context, order *model.Order) (newOrder *model.Order, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, wrapDBError("failed to begin transaction", "", err)
	}

	defer func() {
		err = r.finishTransaction(tx, err)
	}()

	if _, err := r.getOrderByOrderUID(ctx, tx, order.OrderUID); err == nil {
		updatedOrder, err := r.orderFullUpdate(ctx, tx, order)
		if err != nil {
			return nil, wrapDBError("failed to update existing order", order.OrderUID, err)
		}
		return updatedOrder, nil
	}

	if err != nil && !errors.Is(err, srvcerrors.ErrNotFound) {
		return nil, wrapDBError("failed to check existence of order", order.OrderUID, err)
	}

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
		return nil, wrapDBError("failed to insert into orders while creating new order", "", err)
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
		return nil, wrapDBError("failed to insert into deliveries while creating order", "", err)
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
		return nil, wrapDBError("failed to insert into payments while creating order", "", err)
	}
	newOrder.Payment = order.Payment

	newOrder.Items = make([]*model.Item, len(order.Items))
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

		var newItemID int
		if err := row.Scan(&newItemID); err != nil {
			return nil, wrapDBError("failed to insert into items while creating order", "", err)
		}
		newOrder.Items[i] = copyItem(item, newItemID)
	}

	return newOrder, nil
}

func copyItem(item *model.Item, id int) *model.Item {
	return &model.Item{
		ID:          id,
		OrderUID:    item.OrderUID,
		ChrtID:      item.ChrtID,
		TrackNumber: item.TrackNumber,
		Price:       item.Price,
		RID:         item.RID,
		Name:        item.Name,
		Sale:        item.Sale,
		Size:        item.Size,
		TotalPrice:  item.TotalPrice,
		NmID:        item.NmID,
		Brand:       item.Brand,
		Status:      item.Status,
	}
}

func (r *OrderRepository) orderFullUpdate(ctx context.Context, q Querier, o *model.Order) (*model.Order, error) {
	var newOrder *model.Order

	if _, err := q.ExecContext(ctx,
		deleteItemsQuery, o.OrderUID); err != nil {
		return nil, err
	}

	newOrder, err := r.updateOrderFields(ctx, q, o)
	if err != nil {
		return nil, err
	}

	newDelivery, err := r.updateDeliveryFields(ctx, q, &o.Delivery)
	if err != nil {
		return nil, err
	}

	newPayment, err := r.updatePaymentFields(ctx, q, &o.Payment)
	if err != nil {
		return nil, err
	}

	newOrder.Delivery = *newDelivery
	newOrder.Payment = *newPayment

	for _, item := range o.Items {
		newItem, err := r.updateItemFields(ctx, q, item)
		if err != nil {
			return nil, err
		}
		newOrder.Items = append(newOrder.Items, newItem)
	}
	return newOrder, nil
}

func (r *OrderRepository) updateOrderFields(ctx context.Context, q Querier, o *model.Order) (*model.Order, error) {
	row := q.QueryRowContext(ctx, updateOrderQuery,
		o.TrackNumber,
		o.Entry,
		o.Locale,
		o.InternalSignature,
		o.CustomerID,
		o.DeliveryService,
		o.Shardkey,
		o.SmID,
		o.DateCreated,
		o.OofShard,
		o.OrderUID,
	)
	return dto.ScanOrderFromRow(row)
}

func (r *OrderRepository) updateDeliveryFields(ctx context.Context, q Querier, d *model.Delivery) (*model.Delivery, error) {
	row := q.QueryRowContext(ctx, updateDeliveryQuery,
		d.Name,
		d.Phone,
		d.Zip,
		d.City,
		d.Address,
		d.Region,
		d.Email,
		d.OrderUID,
	)
	return dto.ScanDeliveryFromRow(row)
}

func (r *OrderRepository) updatePaymentFields(ctx context.Context, q Querier, p *model.Payment) (*model.Payment, error) {
	row := q.QueryRowContext(ctx, updatePaymentQuery,
		p.RequestID,
		p.Currency,
		p.Provider,
		p.Amount,
		p.PaymentDT,
		p.Bank,
		p.DeliveryCost,
		p.GoodsTotal,
		p.CustomFee,
		p.Transaction,
	)
	return dto.ScanPaymentFromRow(row)
}

func (r *OrderRepository) updateItemFields(ctx context.Context, q Querier, i *model.Item) (*model.Item, error) {
	row := q.QueryRowContext(ctx, updateItemQuery,
		i.OrderUID,
		i.ChrtID,
		i.TrackNumber,
		i.Price,
		i.RID,
		i.Name,
		i.Sale,
		i.Size,
		i.TotalPrice,
		i.NmID,
		i.Brand,
		i.Status,
		i.ID,
	)
	return dto.ScanItemFromRow(row)
}

func (r *OrderRepository) GetOrderByUID(ctx context.Context, orderUID string) (order *model.Order, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, wrapDBError("failed to begin transaction", "", err)
	}

	defer func() {
		err = r.finishTransaction(tx, err)
	}()

	order, err = r.getOrderByOrderUID(ctx, tx, orderUID)
	if err != nil {
		return nil, wrapDBError("failed to get order by id", orderUID, err)
	}

	delivery, err := r.getDeliveryByOrderUID(ctx, tx, orderUID)
	if err != nil {
		return nil, wrapDBError("failed to get delivery by order_uid", orderUID, err)
	}

	payment, err := r.getPaymentByOrderUID(ctx, tx, orderUID)
	if err != nil {
		return nil, wrapDBError("failed to get payment by order id", orderUID, err)
	}

	order.Delivery = *delivery
	order.Payment = *payment

	return order, nil
}

func (r *OrderRepository) GetAllOrders(ctx context.Context, limit int) (orders []*model.Order, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, wrapDBError("failed to begin transaction", "", err)
	}

	defer func() {
		err = r.finishTransaction(tx, err)
	}()

	orders = make([]*model.Order, 0)
	orders, err = r.getAllOrders(ctx, tx, limit)
	if err != nil {
		return nil, wrapDBError("failed to get all orders", "", err)
	}

	for i, order := range orders {
		delivery, err := r.getDeliveryByOrderUID(ctx, tx, order.OrderUID)
		if err != nil {
			return nil, wrapDBError("failed to get delivery of order", order.OrderUID, err)
		}

		payment, err := r.getPaymentByOrderUID(ctx, tx, order.OrderUID)
		if err != nil {
			return nil, wrapDBError("failed to get payment of order", order.OrderUID, err)
		}

		items, err := r.getAllItemsByOrderUID(ctx, tx, order.OrderUID, limit)
		if err != nil {
			return nil, wrapDBError("failed to get items for order", order.OrderUID, err)
		}

		orders[i].Delivery = *delivery
		orders[i].Payment = *payment
		orders[i].Items = items
	}

	return orders, nil
}

func (r *OrderRepository) GetItemsByOrderUID(ctx context.Context, orderUID string, lastID, limit int) (items []*model.Item, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, wrapDBError("failed to begin transaction", "", err)
	}

	defer func() {
		err = r.finishTransaction(tx, err)
	}()

	items = make([]*model.Item, 0, limit)
	items, err = r.getItemsByOrderUID(ctx, tx, orderUID, lastID, limit)
	if err != nil {
		return nil, wrapDBError("failed to get items of order", orderUID, err)
	}

	return items, nil
}

func (r *OrderRepository) getOrderByOrderUID(ctx context.Context, q Querier, orderUID string) (*model.Order, error) {
	row := q.QueryRowContext(ctx, getOrderByIDQuery, orderUID)
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

func (r *OrderRepository) getDeliveryByOrderUID(ctx context.Context, q Querier, orderUID string) (*model.Delivery, error) {
	row := q.QueryRowContext(ctx, getDeliveryByOrderUIDQuery, orderUID)
	delivery, err := dto.ScanDeliveryFromRow(row)
	if err != nil {
		return nil, err
	}
	return delivery, nil
}

func (r *OrderRepository) getPaymentByOrderUID(ctx context.Context, q Querier, orderUID string) (*model.Payment, error) {
	row := q.QueryRowContext(ctx, getPaymentByOrderUIDQuery, orderUID)
	payment, err := dto.ScanPaymentFromRow(row)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (r *OrderRepository) getItemsByOrderUID(ctx context.Context, q Querier, orderUID string, lastID int, limit int) ([]*model.Item, error) {
	var items []*model.Item

	rows, err := q.QueryContext(ctx, getItemsByOrderUIDQuery, orderUID, lastID, limit)
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

func (r *OrderRepository) getAllItemsByOrderUID(ctx context.Context, q Querier, orderUID string, limit int) ([]*model.Item, error) {
	var items []*model.Item

	rows, err := q.QueryContext(ctx, getAllItemsByOrderUIDQuery, orderUID, limit)
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
			return fmt.Errorf("%v\n%w: rollback failed:\n[%v]\n", origErr, srvcerrors.ErrDatabase, rbErr)
		}
		return fmt.Errorf("%w\nrollback succeeded", origErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("%w: commit failed:\n[%v]\n", srvcerrors.ErrDatabase, commitErr)
	}
	return nil
}

func wrapDBError(baseMsg string, id string, err error) error {
	var errType error
	if errors.Is(err, sql.ErrNoRows) {
		errType = srvcerrors.ErrNotFound
	} else {
		errType = srvcerrors.ErrDatabase
	}

	if id != "" {
		return fmt.Errorf("%w: %s %s:\n[%v]\n", errType, baseMsg, id, err)
	}
	return fmt.Errorf("%w: %s:\n[%v]\n", errType, baseMsg, err)
}
