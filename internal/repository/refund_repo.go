package repository

import (
	"ecommerce-backend/internal/model"

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
	return r.db.Create(refund).Error
}

func (r *refundRepository) GetByID(id uint) (*model.Refund, error) {
	var refund model.Refund
	err := r.db.First(&refund, id).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) GetByUserID(userID uint, query *model.RefundListQuery) ([]model.Refund, int64, error) {
	var refunds []model.Refund
	var total int64

	db := r.db.Model(&model.Refund{}).Where("user_id = ?", userID)

	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	db.Count(&total)

	offset := (query.Page - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Order("id DESC").Find(&refunds).Error
	if err != nil {
		return nil, 0, err
	}

	return refunds, total, nil
}

func (r *refundRepository) Update(refund *model.Refund) error {
	return r.db.Save(refund).Error
}
