package repository

import (
	"ecommerce-backend/internal/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *model.Order, items []model.OrderItem) error
	GetByID(id uint) (*model.Order, error)
	GetByOrderNo(orderNo string) (*model.Order, error)
	GetByUserID(userID uint, query *model.OrderListQuery) ([]model.Order, int64, error)
	Update(order *model.Order) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order, items []model.OrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByUserID(userID uint, query *model.OrderListQuery) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	db := r.db.Model(&model.Order{}).Where("user_id = ?", userID)

	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	db.Count(&total)

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Items").Offset(offset).Limit(query.PageSize).Order("id DESC").Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}
