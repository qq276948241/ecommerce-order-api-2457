package repository

import (
	"ecommerce-backend/internal/model"
	"ecommerce-backend/pkg/dbutil"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *model.Payment) error
	GetByID(id uint) (*model.Payment, error)
	GetByOrderID(orderID uint) (*model.Payment, error)
	GetByPaymentNo(paymentNo string) (*model.Payment, error)
	Update(payment *model.Payment) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *model.Payment) error {
	return dbutil.Create(r.db, payment)
}

func (r *paymentRepository) GetByID(id uint) (*model.Payment, error) {
	return dbutil.GetByID[model.Payment](r.db, id)
}

func (r *paymentRepository) GetByOrderID(orderID uint) (*model.Payment, error) {
	return dbutil.GetOne[model.Payment](r.db, []dbutil.Condition{dbutil.Eq("order_id", orderID)})
}

func (r *paymentRepository) GetByPaymentNo(paymentNo string) (*model.Payment, error) {
	return dbutil.GetOne[model.Payment](r.db, []dbutil.Condition{dbutil.Eq("payment_no", paymentNo)})
}

func (r *paymentRepository) Update(payment *model.Payment) error {
	return dbutil.Update(r.db, payment)
}
