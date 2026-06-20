package model

import "time"

const DefaultStockWarningThreshold = 10

type Product struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	Name                 string    `json:"name" gorm:"type:varchar(200);not null"`
	Description          string    `json:"description" gorm:"type:text"`
	Price                float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	Stock                int       `json:"stock" gorm:"not null;default:0"`
	StockWarningThreshold int      `json:"stock_warning_threshold" gorm:"not null;default:10"`
	ImageURL             string    `json:"image_url" gorm:"type:varchar(500)"`
	Category             string    `json:"category" gorm:"type:varchar(100);index"`
	Status               int       `json:"status" gorm:"not null;default:1"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	LowStock             bool      `json:"low_stock" gorm:"-"`
}

type CreateProductRequest struct {
	Name                  string  `json:"name" binding:"required,max=200"`
	Description           string  `json:"description"`
	Price                 float64 `json:"price" binding:"required,gt=0"`
	Stock                 int     `json:"stock" binding:"required,gte=0"`
	StockWarningThreshold  *int    `json:"stock_warning_threshold" binding:"omitempty,gte=0"`
	ImageURL              string  `json:"image_url"`
	Category              string  `json:"category"`
}

type UpdateProductRequest struct {
	Name                  string  `json:"name" binding:"omitempty,max=200"`
	Description           string  `json:"description"`
	Price                 float64 `json:"price" binding:"omitempty,gt=0"`
	Stock                 int     `json:"stock" binding:"omitempty,gte=0"`
	StockWarningThreshold  *int    `json:"stock_warning_threshold" binding:"omitempty,gte=0"`
	ImageURL              string  `json:"image_url"`
	Category              string  `json:"category"`
	Status                *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

type ProductListQuery struct {
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=10"`
	Keyword   string `form:"keyword"`
	Category  string `form:"category"`
	LowStock  *bool  `form:"low_stock"`
}

type ProductListResponse struct {
	List  []Product `json:"list"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}
