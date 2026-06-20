package model

import "time"

const (
	OrderStatusPending   = 0
	OrderStatusPaid      = 1
	OrderStatusShipped   = 2
	OrderStatusCompleted = 3
	OrderStatusCancelled = 4
	OrderStatusRefunded  = 5
)

type Order struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	OrderNo    string    `json:"order_no" gorm:"type:varchar(32);uniqueIndex;not null"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	TotalAmount float64  `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	Status     int       `json:"status" gorm:"not null;default:0"`
	Address    string    `json:"address" gorm:"type:varchar(500)"`
	Remark     string    `json:"remark" gorm:"type:varchar(500)"`
	PayTime    *time.Time `json:"pay_time"`
	Items      []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID         uint    `json:"id" gorm:"primaryKey"`
	OrderID    uint    `json:"order_id" gorm:"not null;index"`
	ProductID  uint    `json:"product_id" gorm:"not null"`
	ProductName string  `json:"product_name" gorm:"type:varchar(200);not null"`
	Price      float64 `json:"price" gorm:"type:decimal(10,2);not null"`
	Quantity   int     `json:"quantity" gorm:"not null"`
	Subtotal   float64 `json:"subtotal" gorm:"type:decimal(10,2);not null"`
}

type CreateOrderRequest struct {
	CartIDs []uint `json:"cart_ids" binding:"required"`
	Address string `json:"address" binding:"required"`
	Remark  string `json:"remark"`
}

type OrderListQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=10"`
	Status   *int `form:"status"`
}

type OrderListResponse struct {
	List  []Order `json:"list"`
	Total int64   `json:"total"`
	Page  int     `json:"page"`
	Size  int     `json:"size"`
}
