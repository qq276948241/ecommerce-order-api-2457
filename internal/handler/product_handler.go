package handler

import (
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
	"ecommerce-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req model.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	product, err := h.productService.Create(&req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, product)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}

	product, err := h.productService.GetByID(id)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}

	var req model.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	product, err := h.productService.Update(id, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}

	if err := h.productService.Delete(id); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ProductHandler) List(c *gin.Context) {
	var query model.ProductListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数错误: "+err.Error())
		return
	}

	result, err := h.productService.List(&query)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, result)
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
