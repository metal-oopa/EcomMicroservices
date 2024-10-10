package models

import "time"

type Order struct {
	OrderID     int
	UserID      int
	Items       []OrderItem
	TotalAmount float64
	Status      string
	CreatedAt   time.Time
}

type OrderItem struct {
	OrderID   int
	ProductID int
	Quantity  int32
}
