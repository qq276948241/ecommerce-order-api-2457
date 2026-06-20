package handler

import (
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
	"ecommerce-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	order, err := h.orderService.CreateOrder(userID, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "订单ID格式错误")
		return
	}

	order, err := h.orderService.GetOrderByID(userID, uint(orderID))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, order)
}

func (h *OrderHandler) GetOrderList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	var query model.OrderListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数错误: "+err.Error())
		return
	}

	result, err := h.orderService.GetOrderList(userID, &query)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "订单ID格式错误")
		return
	}

	err = h.orderService.CancelOrder(userID, uint(orderID))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
