package service

import (
	"fmt"
	"time"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	bizerr "ecommerce-backend/pkg/errors"

	"gorm.io/gorm"
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
	order, err := s.orderRepo.GetByID(req.OrderID)
	if err != nil {
		return nil, bizerr.NotFound("订单不存在")
	}
	if order.UserID != userID {
		return nil, bizerr.Forbidden("无权支付此订单")
	}
	if order.Status != model.OrderStatusPending {
		return nil, bizerr.BadRequest("订单状态不允许支付")
	}

	existingPayment, _ := s.paymentRepo.GetByOrderID(req.OrderID)
	if existingPayment != nil && existingPayment.Status == model.PaymentStatusSuccess {
		return nil, bizerr.BadRequest("订单已支付")
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

	if existingPayment != nil {
		payment.ID = existingPayment.ID
		payment.PaymentNo = existingPayment.PaymentNo
		err = s.paymentRepo.Update(payment)
	} else {
		err = s.paymentRepo.Create(payment)
	}
	if err != nil {
		return nil, bizerr.WrapInternal(err, "创建支付记录失败")
	}

	if err := s.mockPayment(payment); err != nil {
		return nil, err
	}

	return &model.PayResponse{
		PaymentID: payment.ID,
		PaymentNo: payment.PaymentNo,
		Amount:    payment.Amount,
		PayStatus: payment.Status,
	}, nil
}

func (s *paymentService) mockPayment(payment *model.Payment) error {
	now := time.Now()

	err := s.db.Transaction(func(tx *gorm.DB) error {
		payment.Status = model.PaymentStatusSuccess
		payment.TransactionID = fmt.Sprintf("MOCK%d", now.Unix())
		payment.PaidAt = &now
		if err := tx.Save(payment).Error; err != nil {
			return err
		}

		order, err := s.orderRepo.GetByID(payment.OrderID)
		if err != nil {
			return err
		}
		order.Status = model.OrderStatusPaid
		order.PayTime = &now
		if err := tx.Save(order).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return bizerr.WrapInternal(err, "支付处理失败")
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
