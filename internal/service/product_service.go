package service

import (
	"errors"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

type ProductService interface {
	Create(req *model.CreateProductRequest) (*model.Product, error)
	GetByID(id uint) (*model.Product, error)
	Update(id uint, req *model.UpdateProductRequest) (*model.Product, error)
	Delete(id uint) error
	List(query *model.ProductListQuery) (*model.ProductListResponse, error)
}

type productService struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) ProductService {
	return &productService{
		productRepo: productRepo,
	}
}

func markLowStock(product *model.Product) {
	product.LowStock = product.Stock <= product.StockWarningThreshold
}

func markLowStockList(products []model.Product) {
	for i := range products {
		markLowStock(&products[i])
	}
}

func (s *productService) Create(req *model.CreateProductRequest) (*model.Product, error) {
	threshold := model.DefaultStockWarningThreshold
	if req.StockWarningThreshold != nil {
		threshold = *req.StockWarningThreshold
	}

	product := &model.Product{
		Name:                  req.Name,
		Description:           req.Description,
		Price:                 req.Price,
		Stock:                 req.Stock,
		StockWarningThreshold: threshold,
		ImageURL:              req.ImageURL,
		Category:              req.Category,
		Status:                1,
	}

	err := s.productRepo.Create(product)
	if err != nil {
		return nil, errors.New("创建商品失败")
	}

	markLowStock(product)
	return product, nil
}

func (s *productService) GetByID(id uint) (*model.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("商品不存在")
	}
	markLowStock(product)
	return product, nil
}

func (s *productService) Update(id uint, req *model.UpdateProductRequest) (*model.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}
	if req.StockWarningThreshold != nil {
		product.StockWarningThreshold = *req.StockWarningThreshold
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Status != nil {
		product.Status = *req.Status
	}

	err = s.productRepo.Update(product)
	if err != nil {
		return nil, errors.New("更新商品失败")
	}

	markLowStock(product)
	return product, nil
}

func (s *productService) Delete(id uint) error {
	_, err := s.productRepo.GetByID(id)
	if err != nil {
		return errors.New("商品不存在")
	}

	return s.productRepo.Delete(id)
}

func (s *productService) List(query *model.ProductListQuery) (*model.ProductListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 10
	}

	products, total, err := s.productRepo.List(query)
	if err != nil {
		return nil, errors.New("获取商品列表失败")
	}

	markLowStockList(products)

	return &model.ProductListResponse{
		List:  products,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}
