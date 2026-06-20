package repository

import (
	"ecommerce-backend/internal/model"
	"ecommerce-backend/pkg/dbutil"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	return dbutil.Create(r.db, user)
}

func (r *userRepository) GetByID(id uint) (*model.User, error) {
	return dbutil.GetByID[model.User](r.db, id)
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	return dbutil.GetOne[model.User](r.db, []dbutil.Condition{dbutil.Eq("username", username)})
}

func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	return dbutil.GetOne[model.User](r.db, []dbutil.Condition{dbutil.Eq("email", email)})
}
