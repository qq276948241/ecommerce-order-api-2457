package handler

import (
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
	"ecommerce-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, user)
}
