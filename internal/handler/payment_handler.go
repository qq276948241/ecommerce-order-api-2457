package handler

import (
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
	"ecommerce-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentHandler) Pay(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	var req model.PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.paymentService.Pay(userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, result)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	paymentID, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "支付记录ID格式错误")
		return
	}

	payment, err := h.paymentService.GetPaymentByID(userID, paymentID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, payment)
}

func (h *PaymentHandler) GetByOrderID(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	orderID, err := parseUintParam(c, "order_id")
	if err != nil {
		response.BadRequest(c, "订单ID格式错误")
		return
	}

	payment, err := h.paymentService.GetPaymentByOrderID(userID, orderID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, payment)
}
