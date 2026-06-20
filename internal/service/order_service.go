package service

import (
	"fmt"
	"time"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	bizerr "ecommerce-backend/pkg/errors"

	"gorm.io/gorm"
)

type OrderService interface {
	CreateOrder(userID uint, req *model.CreateOrderRequest) (*model.Order, error)
	GetOrderByID(userID, orderID uint) (*model.Order, error)
	GetOrderList(userID uint, query *model.OrderListQuery) (*model.OrderListResponse, error)
	CancelOrder(userID, orderID uint) error
}

type orderService struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	db          *gorm.DB
}

func NewOrderService(orderRepo repository.OrderRepository, cartRepo repository.CartRepository, productRepo repository.ProductRepository, db *gorm.DB) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		db:          db,
	}
}

func (s *orderService) CreateOrder(userID uint, req *model.CreateOrderRequest) (*model.Order, error) {
	if len(req.CartIDs) == 0 {
		return nil, bizerr.BadRequest("请选择要下单的商品")
	}

	cartItems, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return nil, bizerr.WrapInternal(err, "获取购物车失败")
	}

	var selectedItems []model.CartItem
	for _, item := range cartItems {
		for _, id := range req.CartIDs {
			if item.ID == id {
				selectedItems = append(selectedItems, item)
				break
			}
		}
	}

	if len(selectedItems) == 0 {
		return nil, bizerr.NotFound("未找到选中的商品")
	}

	var totalAmount float64
	var orderItems []model.OrderItem

	for _, item := range selectedItems {
		product := item.Product
		if product.Status != 1 {
			return nil, bizerr.BadRequestf("商品 %s 已下架", product.Name)
		}
		if product.Stock < item.Quantity {
			return nil, bizerr.BadRequestf("商品 %s 库存不足", product.Name)
		}

		subtotal := product.Price * float64(item.Quantity)
		totalAmount += subtotal

		orderItems = append(orderItems, model.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       product.Price,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	orderNo := generateOrderNo(userID)

	order := &model.Order{
		OrderNo:     orderNo,
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      model.OrderStatusPending,
		Address:     req.Address,
		Remark:      req.Remark,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		if len(orderItems) > 0 {
			if err := tx.Create(&orderItems).Error; err != nil {
				return err
			}
		}

		for _, item := range selectedItems {
			result := tx.Model(&model.Product{}).
				Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return bizerr.BadRequest("库存扣减失败")
			}
		}

		if err := tx.Where("id IN ?", req.CartIDs).Delete(&model.CartItem{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if _, ok := bizerr.IsBizError(err); ok {
			return nil, err
		}
		return nil, bizerr.WrapInternal(err, "创建订单失败")
	}

	order.Items = orderItems
	return order, nil
}

func (s *orderService) GetOrderByID(userID, orderID uint) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, bizerr.NotFound("订单不存在")
	}
	if order.UserID != userID {
		return nil, bizerr.Forbidden("无权查看此订单")
	}
	return order, nil
}

func (s *orderService) GetOrderList(userID uint, query *model.OrderListQuery) (*model.OrderListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 10
	}

	orders, total, err := s.orderRepo.GetByUserID(userID, query)
	if err != nil {
		return nil, bizerr.WrapInternal(err, "获取订单列表失败")
	}

	return &model.OrderListResponse{
		List:  orders,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *orderService) CancelOrder(userID, orderID uint) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return bizerr.NotFound("订单不存在")
	}
	if order.UserID != userID {
		return bizerr.Forbidden("无权操作此订单")
	}
	if order.Status != model.OrderStatusPending {
		return bizerr.BadRequest("仅待支付订单可取消")
	}

	order.Status = model.OrderStatusCancelled
	if err := s.orderRepo.Update(order); err != nil {
		return bizerr.WrapInternal(err, "取消订单失败")
	}
	return nil
}

func generateOrderNo(userID uint) string {
	now := time.Now()
	return fmt.Sprintf("ORD%d%04d%02d%02d%02d%02d%02d%06d",
		userID,
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		now.Nanosecond()/1000%1000000)
}
