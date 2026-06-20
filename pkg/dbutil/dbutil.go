package dbutil

import (
	"strings"

	"gorm.io/gorm"
)

type Condition struct {
	Column   string
	Operator string
	Value    interface{}
}

type PageQuery struct {
	Page     int
	PageSize int
}

type PageResult[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

func Eq(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: "=", Value: value}
}

func Neq(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: "!=", Value: value}
}

func Gt(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: ">", Value: value}
}

func Gte(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: ">=", Value: value}
}

func Lt(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: "<", Value: value}
}

func Lte(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: "<=", Value: value}
}

func Like(column string, value string) Condition {
	return Condition{Column: column, Operator: "LIKE", Value: "%" + value + "%"}
}

func In(column string, values interface{}) Condition {
	return Condition{Column: column, Operator: "IN", Value: values}
}

func RawExpr(column string, value interface{}) Condition {
	return Condition{Column: column, Operator: "EXPR", Value: value}
}

func ApplyConditions(db *gorm.DB, conditions []Condition) *gorm.DB {
	for _, c := range conditions {
		switch strings.ToUpper(c.Operator) {
		case "IN":
			db = db.Where(c.Column+" IN ?", c.Value)
		case "EXPR":
			db = db.Where(c.Column, c.Value)
		case "LIKE":
			db = db.Where(c.Column+" LIKE ?", c.Value)
		default:
			db = db.Where(c.Column+" "+c.Operator+" ?", c.Value)
		}
	}
	return db
}

func (p *PageQuery) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

func (p *PageQuery) Offset() int {
	return (p.Page - 1) * p.PageSize
}

type QueryBuilder struct {
	db         *gorm.DB
	model      interface{}
	conditions []Condition
	orderBy    []string
}

func NewQuery(db *gorm.DB, model interface{}) *QueryBuilder {
	return &QueryBuilder{
		db:      db.Model(model),
		model:   model,
		orderBy: []string{"id DESC"},
	}
}

func (q *QueryBuilder) Where(conditions ...Condition) *QueryBuilder {
	q.conditions = append(q.conditions, conditions...)
	return q
}

func (q *QueryBuilder) OrderBy(clauses ...string) *QueryBuilder {
	q.orderBy = clauses
	return q
}

func (q *QueryBuilder) DB() *gorm.DB {
	db := ApplyConditions(q.db, q.conditions)
	for _, clause := range q.orderBy {
		db = db.Order(clause)
	}
	return db
}

func Create[T any](db *gorm.DB, entity *T) error {
	return db.Create(entity).Error
}

func GetByID[T any](db *gorm.DB, id uint, preloads ...string) (*T, error) {
	var entity T
	q := db
	for _, p := range preloads {
		q = q.Preload(p)
	}
	err := q.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func GetOne[T any](db *gorm.DB, conditions []Condition, preloads ...string) (*T, error) {
	var entity T
	q := db
	for _, p := range preloads {
		q = q.Preload(p)
	}
	q = ApplyConditions(q.Model(new(T)), conditions)
	err := q.First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func List[T any](db *gorm.DB, conditions []Condition, orderBy []string, preloads ...string) ([]T, error) {
	var list []T
	q := db
	for _, p := range preloads {
		q = q.Preload(p)
	}
	q = ApplyConditions(q.Model(new(T)), conditions)
	for _, clause := range orderBy {
		q = q.Order(clause)
	}
	err := q.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func Paginate[T any](db *gorm.DB, pq PageQuery, conditions []Condition, orderBy []string, preloads ...string) (*PageResult[T], error) {
	pq.Normalize()

	var total int64
	q := ApplyConditions(db.Model(new(T)), conditions)
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	var list []T
	q = db
	for _, p := range preloads {
		q = q.Preload(p)
	}
	q = ApplyConditions(q.Model(new(T)), conditions)
	for _, clause := range orderBy {
		q = q.Order(clause)
	}
	err := q.Offset(pq.Offset()).Limit(pq.PageSize).Find(&list).Error
	if err != nil {
		return nil, err
	}

	return &PageResult[T]{
		List:  list,
		Total: total,
		Page:  pq.Page,
		Size:  pq.PageSize,
	}, nil
}

func Update[T any](db *gorm.DB, entity *T) error {
	return db.Save(entity).Error
}

func UpdateFields(db *gorm.DB, model interface{}, conditions []Condition, fields map[string]interface{}) (int64, error) {
	q := ApplyConditions(db.Model(model), conditions)
	result := q.Updates(fields)
	return result.RowsAffected, result.Error
}

func DeleteByID(db *gorm.DB, model interface{}, id uint) error {
	return db.Delete(model, id).Error
}

func DeleteByConditions(db *gorm.DB, model interface{}, conditions []Condition) error {
	q := ApplyConditions(db.Model(model), conditions)
	return q.Delete(model).Error
}

func Count(db *gorm.DB, model interface{}, conditions []Condition) (int64, error) {
	var count int64
	q := ApplyConditions(db.Model(model), conditions)
	err := q.Count(&count).Error
	return count, err
}

func Exists(db *gorm.DB, model interface{}, conditions []Condition) (bool, error) {
	count, err := Count(db, model, conditions)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
