package handler

import (
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
	"ecommerce-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type RefundHandler struct {
	refundService service.RefundService
}

func NewRefundHandler(refundService service.RefundService) *RefundHandler {
	return &RefundHandler{
		refundService: refundService,
	}
}

func (h *RefundHandler) CreateRefund(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	var req model.CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	refund, err := h.refundService.CreateRefund(userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, refund)
}

func (h *RefundHandler) GetRefund(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	refundID, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "退款记录ID格式错误")
		return
	}

	refund, err := h.refundService.GetRefundByID(userID, refundID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, refund)
}

func (h *RefundHandler) GetRefundList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "用户未认证")
		return
	}

	var query model.RefundListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数错误: "+err.Error())
		return
	}

	result, err := h.refundService.GetRefundList(userID, &query)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, result)
}
