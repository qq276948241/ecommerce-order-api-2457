package model

import "time"

const (
	PaymentStatusPending = 0
	PaymentStatusSuccess = 1
	PaymentStatusFailed  = 2
)

type Payment struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	PaymentNo     string    `json:"payment_no" gorm:"type:varchar(32);uniqueIndex;not null"`
	OrderID       uint      `json:"order_id" gorm:"not null;index"`
	UserID        uint      `json:"user_id" gorm:"not null;index"`
	Amount        float64   `json:"amount" gorm:"type:decimal(10,2);not null"`
	PaymentMethod string    `json:"payment_method" gorm:"type:varchar(50)"`
	Status        int       `json:"status" gorm:"not null;default:0"`
	TransactionID string    `json:"transaction_id" gorm:"type:varchar(100)"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PayRequest struct {
	OrderID       uint   `json:"order_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

type PayResponse struct {
	PaymentID   uint   `json:"payment_id"`
	PaymentNo   string `json:"payment_no"`
	Amount      float64 `json:"amount"`
	PayStatus   int    `json:"pay_status"`
	PaymentURL  string `json:"payment_url,omitempty"`
}
