package repository

import (
	"ecommerce-backend/internal/model"
	"ecommerce-backend/pkg/dbutil"

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
	return dbutil.Create(r.db, item)
}

func (r *cartRepository) GetByUserID(userID uint) ([]model.CartItem, error) {
	return dbutil.List[model.CartItem](r.db,
		[]dbutil.Condition{dbutil.Eq("user_id", userID)},
		[]string{"id DESC"},
		"Product")
}

func (r *cartRepository) GetByID(id uint) (*model.CartItem, error) {
	return dbutil.GetByID[model.CartItem](r.db, id, "Product")
}

func (r *cartRepository) GetByUserAndProduct(userID, productID uint) (*model.CartItem, error) {
	return dbutil.GetOne[model.CartItem](r.db,
		[]dbutil.Condition{
			dbutil.Eq("user_id", userID),
			dbutil.Eq("product_id", productID),
		})
}

func (r *cartRepository) Update(item *model.CartItem) error {
	return dbutil.Update(r.db, item)
}

func (r *cartRepository) Delete(id uint) error {
	return dbutil.DeleteByID(r.db, &model.CartItem{}, id)
}

func (r *cartRepository) DeleteByIDs(ids []uint, userID uint) error {
	return dbutil.DeleteByConditions(r.db, &model.CartItem{},
		[]dbutil.Condition{
			dbutil.In("id", ids),
			dbutil.Eq("user_id", userID),
		})
}
