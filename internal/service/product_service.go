package service

import (
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	bizerr "ecommerce-backend/pkg/errors"
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

	if err := s.productRepo.Create(product); err != nil {
		return nil, bizerr.WrapInternal(err, "创建商品失败")
	}

	markLowStock(product)
	return product, nil
}

func (s *productService) GetByID(id uint) (*model.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, bizerr.NotFound("商品不存在")
	}
	markLowStock(product)
	return product, nil
}

func (s *productService) Update(id uint, req *model.UpdateProductRequest) (*model.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, bizerr.NotFound("商品不存在")
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

	if err := s.productRepo.Update(product); err != nil {
		return nil, bizerr.WrapInternal(err, "更新商品失败")
	}

	markLowStock(product)
	return product, nil
}

func (s *productService) Delete(id uint) error {
	_, err := s.productRepo.GetByID(id)
	if err != nil {
		return bizerr.NotFound("商品不存在")
	}

	if err := s.productRepo.Delete(id); err != nil {
		return bizerr.WrapInternal(err, "删除商品失败")
	}
	return nil
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
		return nil, bizerr.WrapInternal(err, "获取商品列表失败")
	}

	markLowStockList(products)

	return &model.ProductListResponse{
		List:  products,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}
