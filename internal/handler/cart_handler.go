package handler

import (
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
	"ecommerce-backend/pkg/response"
	"strconv"

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

	err := h.cartService.AddItem(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
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
		response.Error(c, err.Error())
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

	idStr := c.Param("id")
	itemID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "购物车项ID格式错误")
		return
	}

	var req model.UpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err = h.cartService.UpdateItem(userID, uint(itemID), req.Quantity)
	if err != nil {
		response.Error(c, err.Error())
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

	idStr := c.Param("id")
	itemID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "购物车项ID格式错误")
		return
	}

	err = h.cartService.DeleteItem(userID, uint(itemID))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
