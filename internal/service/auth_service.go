package service

import (
	"errors"

	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req *model.RegisterRequest) (*model.User, error)
	Login(req *model.LoginRequest) (*model.LoginResponse, error)
	GetUserByID(id uint) (*model.User, error)
}

type authService struct {
	userRepo repository.UserRepository
	jwtCfg   *config.JWTConfig
}

func NewAuthService(userRepo repository.UserRepository, jwtCfg *config.JWTConfig) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

func (s *authService) Register(req *model.RegisterRequest) (*model.User, error) {
	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	existingEmail, _ := s.userRepo.GetByEmail(req.Email)
	if existingEmail != nil {
		return nil, errors.New("邮箱已被注册")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Nickname: req.Nickname,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, errors.New("注册失败")
	}

	return user, nil
}

func (s *authService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := middleware.GenerateToken(s.jwtCfg, user.ID, user.Username)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *authService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.GetByID(id)
}
