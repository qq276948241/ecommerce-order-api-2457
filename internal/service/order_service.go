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

	var finalOrder *model.Order
	var finalOrderItems []model.OrderItem

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var cartItems []model.CartItem
		if err := tx.Preload("Product").
			Where("user_id = ? AND id IN ?", userID, req.CartIDs).
			Find(&cartItems).Error; err != nil {
			return err
		}
		if len(cartItems) == 0 {
			return bizerr.NotFound("未找到选中的商品")
		}

		productIDs := make([]uint, 0, len(cartItems))
		for _, item := range cartItems {
			productIDs = append(productIDs, item.ProductID)
		}

		var products []model.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", productIDs).
			Find(&products).Error; err != nil {
			return err
		}

		productMap := make(map[uint]*model.Product, len(products))
		for i := range products {
			productMap[products[i].ID] = &products[i]
		}

		var totalAmount float64
		var orderItems []model.OrderItem

		for _, item := range cartItems {
			product, ok := productMap[item.ProductID]
			if !ok {
				return bizerr.NotFoundf("商品不存在")
			}
			if product.Status != 1 {
				return bizerr.BadRequestf("商品 %s 已下架", product.Name)
			}
			if product.Stock < item.Quantity {
				return bizerr.BadRequestf("商品 %s 库存不足", product.Name)
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

		for _, item := range cartItems {
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

		order.Items = orderItems
		finalOrder = order
		finalOrderItems = orderItems
		return nil
	})

	if err != nil {
		if _, ok := bizerr.IsBizError(err); ok {
			return nil, err
		}
		return nil, bizerr.WrapInternal(err, "创建订单失败")
	}

	finalOrder.Items = finalOrderItems
	return finalOrder, nil
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
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			First(&order, orderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return bizerr.NotFound("订单不存在")
			}
			return err
		}
		if order.UserID != userID {
			return bizerr.Forbidden("无权操作此订单")
		}
		if order.Status != model.OrderStatusPending {
			return bizerr.BadRequest("仅待支付订单可取消")
		}

		result := tx.Model(&model.Order{}).
			Where("id = ? AND status = ?", orderID, model.OrderStatusPending).
			Update("status", model.OrderStatusCancelled)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return bizerr.BadRequest("订单状态已变更，取消失败")
		}

		for _, item := range order.Items {
			if err := tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if _, ok := bizerr.IsBizError(err); ok {
			return err
		}
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
