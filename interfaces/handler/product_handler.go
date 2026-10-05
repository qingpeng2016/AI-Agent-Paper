package handler

import (
	"errors"
	"net/http"
	"strconv"

	productsvc "github.com/qingpeng2016/ai-token-mall/application/core-service/product"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productSvc *productsvc.ProductService
}

func NewProductHandler(productSvc *productsvc.ProductService) *ProductHandler {
	return &ProductHandler{productSvc: productSvc}
}

// List 在售商品列表（首页 catalog）
func (h *ProductHandler) List(c *gin.Context) {
	var categoryID uint
	if raw := c.Query("products_category_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.ResponseErr(c, err)
			return
		}
		categoryID = uint(id)
	}
	data, err := h.productSvc.List(c.Request.Context(), categoryID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// NavMenu 顶部套餐购买 mega 菜单 + 品牌下拉
func (h *ProductHandler) NavMenu(c *gin.Context) {
	data, err := h.productSvc.NavMenu(c.Request.Context())
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// DetailBySlug 商品详情页
func (h *ProductHandler) DetailBySlug(c *gin.Context) {
	slug := c.Param("slug")
	data, err := h.productSvc.DetailBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, productsvc.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "product not found"})
			return
		}
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
