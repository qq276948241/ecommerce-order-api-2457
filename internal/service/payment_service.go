package service

import (
	"errors"
	"fmt"
	"time"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	bizerr "ecommerce-backend/pkg/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentService interface {
	Pay(userID uint, req *model.PayRequest) (*model.PayResponse, error)
	GetPaymentByID(userID, paymentID uint) (*model.Payment, error)
	GetPaymentByOrderID(userID, orderID uint) (*model.Payment, error)
}

type paymentService struct {
	paymentRepo repository.PaymentRepository
	orderRepo   repository.OrderRepository
	db          *gorm.DB
}

func NewPaymentService(paymentRepo repository.PaymentRepository, orderRepo repository.OrderRepository, db *gorm.DB) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		db:          db,
	}
}

func (s *paymentService) Pay(userID uint, req *model.PayRequest) (*model.PayResponse, error) {
	var resp *model.PayResponse

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&order, req.OrderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return bizerr.NotFound("订单不存在")
			}
			return err
		}
		if order.UserID != userID {
			return bizerr.Forbidden("无权支付此订单")
		}
		if order.Status != model.OrderStatusPending {
			return bizerr.BadRequest("订单状态不允许支付")
		}

		var existingPayment model.Payment
		paymentFindErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_id = ?", req.OrderID).
			First(&existingPayment).Error
		if paymentFindErr != nil && !errors.Is(paymentFindErr, gorm.ErrRecordNotFound) {
			return paymentFindErr
		}

		if paymentFindErr == nil && existingPayment.Status == model.PaymentStatusSuccess {
			return bizerr.BadRequest("订单已支付")
		}

		paymentNo := generatePaymentNo(userID)
		payment := &model.Payment{
			PaymentNo:     paymentNo,
			OrderID:       req.OrderID,
			UserID:        userID,
			Amount:        order.TotalAmount,
			PaymentMethod: req.PaymentMethod,
			Status:        model.PaymentStatusPending,
		}

		if paymentFindErr == nil {
			payment.ID = existingPayment.ID
			payment.PaymentNo = existingPayment.PaymentNo
			payment.Status = model.PaymentStatusPending
			if err := tx.Save(payment).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(payment).Error; err != nil {
				return err
			}
		}

		if err := s.mockPaymentInTx(tx, payment, &order); err != nil {
			return err
		}

		resp = &model.PayResponse{
			PaymentID: payment.ID,
			PaymentNo: payment.PaymentNo,
			Amount:    payment.Amount,
			PayStatus: payment.Status,
		}
		return nil
	})

	if err != nil {
		if _, ok := bizerr.IsBizError(err); ok {
			return nil, err
		}
		return nil, bizerr.WrapInternal(err, "支付处理失败")
	}

	return resp, nil
}

func (s *paymentService) mockPaymentInTx(tx *gorm.DB, payment *model.Payment, order *model.Order) error {
	now := time.Now()

	payment.Status = model.PaymentStatusSuccess
	payment.TransactionID = fmt.Sprintf("MOCK%d", now.Unix())
	payment.PaidAt = &now
	if err := tx.Save(payment).Error; err != nil {
		return err
	}

	result := tx.Model(&model.Order{}).
		Where("id = ? AND status = ?", order.ID, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":   model.OrderStatusPaid,
			"pay_time": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return bizerr.BadRequest("订单状态已变更，支付失败")
	}

	return nil
}

func (s *paymentService) GetPaymentByID(userID, paymentID uint) (*model.Payment, error) {
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, bizerr.NotFound("支付记录不存在")
	}
	if payment.UserID != userID {
		return nil, bizerr.Forbidden("无权查看此支付记录")
	}
	return payment, nil
}

func (s *paymentService) GetPaymentByOrderID(userID, orderID uint) (*model.Payment, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, bizerr.NotFound("订单不存在")
	}
	if order.UserID != userID {
		return nil, bizerr.Forbidden("无权查看此订单的支付记录")
	}

	payment, err := s.paymentRepo.GetByOrderID(orderID)
	if err != nil {
		return nil, bizerr.NotFound("支付记录不存在")
	}
	return payment, nil
}

func generatePaymentNo(userID uint) string {
	now := time.Now()
	return fmt.Sprintf("PAY%d%04d%02d%02d%02d%02d%02d%06d",
		userID,
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		now.Nanosecond()/1000%1000000)
}
