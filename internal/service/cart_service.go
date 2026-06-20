package service

import (
	"errors"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

type CartService interface {
	AddItem(userID uint, req *model.AddCartRequest) error
	GetCart(userID uint) (*model.CartListResponse, error)
	UpdateItem(userID, itemID uint, quantity int) error
	DeleteItem(userID, itemID uint) error
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *cartService) AddItem(userID uint, req *model.AddCartRequest) error {
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return errors.New("商品不存在")
	}
	if product.Status != 1 {
		return errors.New("商品已下架")
	}
	if product.Stock < req.Quantity {
		return errors.New("商品库存不足")
	}

	existingItem, err := s.cartRepo.GetByUserAndProduct(userID, req.ProductID)
	if err == nil && existingItem != nil {
		existingItem.Quantity += req.Quantity
		if product.Stock < existingItem.Quantity {
			return errors.New("商品库存不足")
		}
		return s.cartRepo.Update(existingItem)
	}

	cartItem := &model.CartItem{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}

	return s.cartRepo.AddItem(cartItem)
}

func (s *cartService) GetCart(userID uint) (*model.CartListResponse, error) {
	items, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("获取购物车失败")
	}

	var total float64
	for i := range items {
		markLowStock(&items[i].Product)
		total += float64(items[i].Quantity) * items[i].Product.Price
	}

	return &model.CartListResponse{
		List:  items,
		Total: total,
	}, nil
}

func (s *cartService) UpdateItem(userID, itemID uint, quantity int) error {
	item, err := s.cartRepo.GetByID(itemID)
	if err != nil {
		return errors.New("购物车项不存在")
	}
	if item.UserID != userID {
		return errors.New("无权操作此购物车项")
	}

	product, err := s.productRepo.GetByID(item.ProductID)
	if err != nil {
		return errors.New("商品不存在")
	}
	if product.Stock < quantity {
		return errors.New("商品库存不足")
	}

	item.Quantity = quantity
	return s.cartRepo.Update(item)
}

func (s *cartService) DeleteItem(userID, itemID uint) error {
	item, err := s.cartRepo.GetByID(itemID)
	if err != nil {
		return errors.New("购物车项不存在")
	}
	if item.UserID != userID {
		return errors.New("无权操作此购物车项")
	}

	return s.cartRepo.Delete(itemID)
}
