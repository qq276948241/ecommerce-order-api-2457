package repository

import (
	"ecommerce-backend/internal/model"
	"ecommerce-backend/pkg/dbutil"

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
		if err := dbutil.Create(tx, order); err != nil {
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
	return dbutil.GetByID[model.Order](r.db, id, "Items")
}

func (r *orderRepository) GetByOrderNo(orderNo string) (*model.Order, error) {
	return dbutil.GetOne[model.Order](r.db,
		[]dbutil.Condition{dbutil.Eq("order_no", orderNo)})
}

func (r *orderRepository) GetByUserID(userID uint, query *model.OrderListQuery) ([]model.Order, int64, error) {
	var conds []dbutil.Condition
	conds = append(conds, dbutil.Eq("user_id", userID))

	if query.Status != nil {
		conds = append(conds, dbutil.Eq("status", *query.Status))
	}

	pq := dbutil.PageQuery{Page: query.Page, PageSize: query.PageSize}
	result, err := dbutil.Paginate[model.Order](r.db, pq, conds, []string{"id DESC"}, "Items")
	if err != nil {
		return nil, 0, err
	}
	return result.List, result.Total, nil
}

func (r *orderRepository) Update(order *model.Order) error {
	return dbutil.Update(r.db, order)
}
