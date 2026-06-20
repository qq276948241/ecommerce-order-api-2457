package repository

import (
	"ecommerce-backend/internal/model"

	"gorm.io/gorm"
)

type CartRepository interface {
	AddItem(item *model.CartItem) error
	GetByUserID(userID uint) ([]model.CartItem, error)
	GetByID(id uint) (*model.CartItem, error)
	GetByUserAndProduct(userID, productID uint) (*model.CartItem, error)
	Update(item *model.CartItem) error
	Delete(id uint) error
	DeleteByIDs(ids []uint, userID uint) error
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) AddItem(item *model.CartItem) error {
	return r.db.Create(item).Error
}

func (r *cartRepository) GetByUserID(userID uint) ([]model.CartItem, error) {
	var items []model.CartItem
	err := r.db.Where("user_id = ?", userID).Preload("Product").Order("id DESC").Find(&items).Error
	return items, err
}

func (r *cartRepository) GetByID(id uint) (*model.CartItem, error) {
	var item model.CartItem
	err := r.db.Preload("Product").First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) GetByUserAndProduct(userID, productID uint) (*model.CartItem, error) {
	var item model.CartItem
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) Update(item *model.CartItem) error {
	return r.db.Save(item).Error
}

func (r *cartRepository) Delete(id uint) error {
	return r.db.Delete(&model.CartItem{}, id).Error
}

func (r *cartRepository) DeleteByIDs(ids []uint, userID uint) error {
	return r.db.Where("id IN ? AND user_id = ?", ids, userID).Delete(&model.CartItem{}).Error
}
