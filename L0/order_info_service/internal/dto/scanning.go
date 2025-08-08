package dto

import (
	"database/sql"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
)

func ScanOrderFromRow(row *sql.Row) (*model.Order, error){
	var order model.Order
	
	if err := row.Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	); err != nil {
		return nil, err
	}
	
	return &order, nil
}

func ScanDeliveryFromRow(row *sql.Row) (*model.Delivery, error){
	var delivery model.Delivery
	
	if err := row.Scan(
		&delivery.OrderUID,
		&delivery.Name,    
		&delivery.Phone,   
		&delivery.Zip,     
		&delivery.City,    
		&delivery.Address, 
		&delivery.Region,  
		&delivery.Email,   
	); err != nil {
		return nil, err
	}
	
	return &delivery, nil
}

func ScanPaymentFromRow(row *sql.Row) (*model.Payment, error){
	var payment model.Payment
	
	if err := row.Scan(
		&payment.Transaction,   
		&payment.RequestID,     
		&payment.Currency,      
		&payment.Provider,      
		&payment.Amount,       
		&payment.PaymentDT,     
		&payment.Bank,          
		&payment.DeliveryCost,  
		&payment.GoodsTotal,    
		&payment.CustomFee,     
	); err != nil {
		return nil, err
	}
	
	return &payment, nil
}

func ScanItemFromRow(row *sql.Rows) (*model.Item, error){
	var item model.Item
	
	if err := row.Scan(
		&item.ChrtID,      
		&item.TrackNumber, 
		&item.Price,   
		&item.RID,         
		&item.Name,        
		&item.Sale,     
		&item.Size,        
		&item.TotalPrice, 
		&item.NmID,        
		&item.Brand,       
		&item.Status,      
	); err != nil {
		return nil, err
	}
	
	return &item, nil
}