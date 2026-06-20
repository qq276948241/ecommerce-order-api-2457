package repository

import (
	"ecommerce-backend/internal/model"
	"ecommerce-backend/pkg/dbutil"

	"gorm.io/gorm"
)

type RefundRepository interface {
	Create(refund *model.Refund) error
	GetByID(id uint) (*model.Refund, error)
	GetByUserID(userID uint, query *model.RefundListQuery) ([]model.Refund, int64, error)
	Update(refund *model.Refund) error
}

type refundRepository struct {
	db *gorm.DB
}

func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(refund *model.Refund) error {
	return dbutil.Create(r.db, refund)
}

func (r *refundRepository) GetByID(id uint) (*model.Refund, error) {
	return dbutil.GetByID[model.Refund](r.db, id)
}

func (r *refundRepository) GetByUserID(userID uint, query *model.RefundListQuery) ([]model.Refund, int64, error) {
	var conds []dbutil.Condition
	conds = append(conds, dbutil.Eq("user_id", userID))

	if query.Status != nil {
		conds = append(conds, dbutil.Eq("status", *query.Status))
	}

	pq := dbutil.PageQuery{Page: query.Page, PageSize: query.PageSize}
	result, err := dbutil.Paginate[model.Refund](r.db, pq, conds, []string{"id DESC"})
	if err != nil {
		return nil, 0, err
	}
	return result.List, result.Total, nil
}

func (r *refundRepository) Update(refund *model.Refund) error {
	return dbutil.Update(r.db, refund)
}
