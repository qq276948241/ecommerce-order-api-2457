package service

import (
	"fmt"
	"time"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	bizerr "ecommerce-backend/pkg/errors"
)

type RefundService interface {
	CreateRefund(userID uint, req *model.CreateRefundRequest) (*model.Refund, error)
	GetRefundByID(userID, refundID uint) (*model.Refund, error)
	GetRefundList(userID uint, query *model.RefundListQuery) (*model.RefundListResponse, error)
}

type refundService struct {
	refundRepo repository.RefundRepository
	orderRepo  repository.OrderRepository
}

func NewRefundService(refundRepo repository.RefundRepository, orderRepo repository.OrderRepository) RefundService {
	return &refundService{
		refundRepo: refundRepo,
		orderRepo:  orderRepo,
	}
}

func (s *refundService) CreateRefund(userID uint, req *model.CreateRefundRequest) (*model.Refund, error) {
	order, err := s.orderRepo.GetByID(req.OrderID)
	if err != nil {
		return nil, bizerr.NotFound("订单不存在")
	}
	if order.UserID != userID {
		return nil, bizerr.Forbidden("无权操作此订单")
	}
	if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusCompleted {
		return nil, bizerr.BadRequest("仅已支付或已完成订单可申请退款")
	}
	if req.Amount > order.TotalAmount {
		return nil, bizerr.BadRequest("退款金额不能超过订单金额")
	}

	refundNo := generateRefundNo(userID)
	refund := &model.Refund{
		RefundNo: refundNo,
		OrderID:  req.OrderID,
		UserID:   userID,
		Amount:   req.Amount,
		Reason:   req.Reason,
		Status:   model.RefundStatusPending,
	}

	if err := s.refundRepo.Create(refund); err != nil {
		return nil, bizerr.WrapInternal(err, "申请退款失败")
	}

	return refund, nil
}

func (s *refundService) GetRefundByID(userID, refundID uint) (*model.Refund, error) {
	refund, err := s.refundRepo.GetByID(refundID)
	if err != nil {
		return nil, bizerr.NotFound("退款记录不存在")
	}
	if refund.UserID != userID {
		return nil, bizerr.Forbidden("无权查看此退款记录")
	}
	return refund, nil
}

func (s *refundService) GetRefundList(userID uint, query *model.RefundListQuery) (*model.RefundListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 10
	}

	refunds, total, err := s.refundRepo.GetByUserID(userID, query)
	if err != nil {
		return nil, bizerr.WrapInternal(err, "获取退款列表失败")
	}

	return &model.RefundListResponse{
		List:  refunds,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func generateRefundNo(userID uint) string {
	now := time.Now()
	return fmt.Sprintf("REF%d%04d%02d%02d%02d%02d%02d%06d",
		userID,
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		now.Nanosecond()/1000%1000000)
}
