package handler

import (
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
	"ecommerce-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	var req model.AddCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.cartService.AddItem(userID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	result, err := h.cartService.GetCart(userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, result)
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	itemID, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "购物车项ID格式错误")
		return
	}

	var req model.UpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.cartService.UpdateItem(userID, itemID, req.Quantity); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *CartHandler) DeleteItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	itemID, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "购物车项ID格式错误")
		return
	}

	if err := h.cartService.DeleteItem(userID, itemID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
