package model

import "time"

type CartItem struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	ProductID uint      `json:"product_id" gorm:"not null;index"`
	Quantity  int       `json:"quantity" gorm:"not null;default:1"`
	Product   Product   `json:"product" gorm:"foreignKey:ProductID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gte=1"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,gte=1"`
}

type CartListResponse struct {
	List  []CartItem `json:"list"`
	Total float64    `json:"total"`
}
