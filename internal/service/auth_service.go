package service

import (
	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	bizerr "ecommerce-backend/pkg/errors"

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
		return nil, bizerr.Conflict("用户名已存在")
	}

	existingEmail, _ := s.userRepo.GetByEmail(req.Email)
	if existingEmail != nil {
		return nil, bizerr.Conflict("邮箱已被注册")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, bizerr.WrapInternal(err, "密码加密失败")
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Nickname: req.Nickname,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, bizerr.WrapInternal(err, "注册失败")
	}

	return user, nil
}

func (s *authService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, bizerr.BadRequest("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, bizerr.BadRequest("用户名或密码错误")
	}

	token, err := middleware.GenerateToken(s.jwtCfg, user.ID, user.Username)
	if err != nil {
		return nil, bizerr.WrapInternal(err, "生成令牌失败")
	}

	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *authService) GetUserByID(id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, bizerr.NotFound("用户不存在")
	}
	return user, nil
}
