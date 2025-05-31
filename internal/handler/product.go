package handler

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"github.com/gin-gonic/gin"
)

type ProductService interface {
	GetProductByID(ctx context.Context, id string) (entity.Product, error)
	SelectProducts(ctx context.Context, offset, limit int) ([]entity.Product, int, error)
	SelectProductsByName(ctx context.Context, name string, offset, limit int) ([]entity.Product, int, error)
	SelectProductsByFilters(ctx context.Context, categoryID int, filters map[string]interface{},
		limit, offset int) ([]entity.Product, int, error)
	SelectShopProducts(ctx context.Context, storeID int, offset, limit int) ([]entity.Product, int, error)
}

type ProductHandler struct {
	productService ProductService
}

func NewProductHandler(productService ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Query("id")

	product, err := h.productService.GetProductByID(context.Background(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) SelectProducts(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid page number",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid limit value (should be between 1 and 100)",
		})
		return
	}

	offset := (page - 1) * limit

	products, total, err := h.productService.SelectProducts(context.Background(), offset, limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	})
}

func (h *ProductHandler) SearchProductsByName(c *gin.Context) {
	name := c.Query("name")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid page number",
		})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid limit value (should be between 1 and 100)",
		})
		return
	}

	offset := (page - 1) * limit

	products, total, err := h.productService.SelectProductsByName(c.Request.Context(), name, offset, limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	})
}

func (h *ProductHandler) SelectFilteredProducts(c *gin.Context) {
	categoryIDStr := c.Query("category_id")
	if categoryIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category_id is required"})
		return
	}
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	filters := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key == "category_id" || key == "limit" || key == "offset" {
			continue
		}
		if len(values) > 0 {
			filters[key] = values[0]
		}
	}
	offset := (page - 1) * limit

	products, tot, err := h.productService.SelectProductsByFilters(c.Request.Context(), categoryID, filters, offset, limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	totalPages := int(math.Ceil(float64(tot) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_items":  tot,
			"total_pages":  totalPages,
		},
	})
}

func (h *ProductHandler) SelectShopProducts(c *gin.Context) {
	shopIDStr := c.Query("shop_id")
	if shopIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shop_id is required"})
		return
	}
	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop_id"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid page number",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid limit value (should be between 1 and 100)",
		})
		return
	}

	offset := (page - 1) * limit

	products, total, err := h.productService.SelectShopProducts(context.Background(), shopID, offset, limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	})
}