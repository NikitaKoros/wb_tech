package model

import "time"

type Order struct {
	OrderUID          string    `json:"order_uid" validate:"required,alphanum"`
	TrackNumber       string    `json:"track_number" validate:"required"`
	Entry             string    `json:"entry" validate:"required"`
	Delivery          Delivery  `json:"delivery" validate:"required,dive"`
	Payment           Payment   `json:"payment" validate:"required,dive"`
	Items             []Item    `json:"items" validate:"required,min=1,dive"`
	Locale            string    `json:"locale" validate:"required,oneof=en ru"`
	InternalSignature string    `json:"internal_signature" validate:"omitempty"`
	CustomerID        string    `json:"customer_id" validate:"required"`
	DeliveryService   string    `json:"delivery_service" validate:"required"`
	Shardkey          string    `json:"shardkey" validate:"required"`
	SmID              int       `json:"sm_id" validate:"required"`
	DateCreated       time.Time `json:"date_created" validate:"required"`
	OofShard          string    `json:"oof_shard" validate:"required"`
}

type Delivery struct {
	OrderUID string `json:"order_uid" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone" validate:"required,e164"`
	Zip      string `json:"zip" validate:"required,numeric"`
	City     string `json:"city" validate:"required"`
	Address  string `json:"address" validate:"required"`
	Region   string `json:"region" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type Payment struct {
	OrderUID string `json:"order_uid" validate:"required"`
	Transaction  string `json:"transaction" validate:"required"`
	RequestID    string `json:"request_id" validate:"omitempty"`
	Currency     string `json:"currency" validate:"required,oneof=USD RUB"`
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"required,min=0"`
	PaymentDT    int    `json:"payment_dt" validate:"required"`
	Bank         string `json:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"required,min=0"`
	GoodsTotal   int    `json:"goods_total" validate:"required,min=0"`
	CustomFee    int    `json:"custom_fee" validate:"required,min=0"`
}

type Item struct {
	OrderUID string `json:"order_uid" validate:"required"`
	ID          int    `json:"id" validate:"omitempty"`
	ChrtID      int    `json:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price" validate:"required,min=0"`
	RID         string `json:"rid" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale" validate:"required,min=0,max=100"`
	Size        string `json:"size" validate:"required"`
	TotalPrice  int    `json:"total_price" validate:"required,min=0"`
	NmID        int    `json:"nm_id" validate:"required"`
	Brand       string `json:"brand" validate:"required"`
	Status      int    `json:"status" validate:"required"`
}
