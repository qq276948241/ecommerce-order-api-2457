package model

import "time"

const (
	RefundStatusPending  = 0
	RefundStatusApproved = 1
	RefundStatusRejected = 2
	RefundStatusCompleted = 3
)

type Refund struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	RefundNo      string    `json:"refund_no" gorm:"type:varchar(32);uniqueIndex;not null"`
	OrderID       uint      `json:"order_id" gorm:"not null;index"`
	UserID        uint      `json:"user_id" gorm:"not null;index"`
	Amount        float64   `json:"amount" gorm:"type:decimal(10,2);not null"`
	Reason        string    `json:"reason" gorm:"type:varchar(500);not null"`
	Status        int       `json:"status" gorm:"not null;default:0"`
	Remark        string    `json:"remark" gorm:"type:varchar(500)"`
	RefundedAt    *time.Time `json:"refunded_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateRefundRequest struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Reason  string `json:"reason" binding:"required,max=500"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

type RefundListQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=10"`
	Status   *int `form:"status"`
}

type RefundListResponse struct {
	List  []Refund `json:"list"`
	Total int64    `json:"total"`
	Page  int      `json:"page"`
	Size  int      `json:"size"`
}
