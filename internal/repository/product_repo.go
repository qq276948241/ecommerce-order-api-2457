package repository

import (
	"ecommerce-backend/internal/model"
	"ecommerce-backend/pkg/dbutil"

	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *model.Product) error
	GetByID(id uint) (*model.Product, error)
	Update(product *model.Product) error
	Delete(id uint) error
	List(query *model.ProductListQuery) ([]model.Product, int64, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *model.Product) error {
	return dbutil.Create(r.db, product)
}

func (r *productRepository) GetByID(id uint) (*model.Product, error) {
	return dbutil.GetByID[model.Product](r.db, id)
}

func (r *productRepository) Update(product *model.Product) error {
	return dbutil.Update(r.db, product)
}

func (r *productRepository) Delete(id uint) error {
	return dbutil.DeleteByID(r.db, &model.Product{}, id)
}

func (r *productRepository) List(query *model.ProductListQuery) ([]model.Product, int64, error) {
	var conds []dbutil.Condition
	conds = append(conds, dbutil.Eq("status", 1))

	if query.Keyword != "" {
		conds = append(conds, dbutil.Like("name", query.Keyword))
	}
	if query.Category != "" {
		conds = append(conds, dbutil.Eq("category", query.Category))
	}
	if query.LowStock != nil && *query.LowStock {
		conds = append(conds, dbutil.RawExpr("stock <= stock_warning_threshold", nil))
	}

	pq := dbutil.PageQuery{Page: query.Page, PageSize: query.PageSize}
	result, err := dbutil.Paginate[model.Product](r.db, pq, conds, []string{"id DESC"})
	if err != nil {
		return nil, 0, err
	}
	return result.List, result.Total, nil
}
