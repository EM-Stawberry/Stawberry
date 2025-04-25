package handler

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product product.Product) (uint, error)
	GetProductByID(ctx context.Context, id string) (entity.Product, error)
	SelectProducts(ctx context.Context, offset, limit int) ([]entity.Product, int, error)
	SelectProductsByName(ctx context.Context, name string, offset, limit int) ([]entity.Product, int, error)
	SelectProductsByCategoryAndAttributes(ctx context.Context, categoryID int, filters map[string]interface{}, limit, offset int,) ([]entity.Product, int, error)
	GetStoreProducts(ctx context.Context, id string, offset, limit int) ([]entity.Product, int, error)
	UpdateProduct(ctx context.Context, id string, updateProduct product.UpdateProduct) error
}

type productHandler struct {
	productService ProductService
}

 func NewProductHandler(productService ProductService) productHandler {
	return productHandler{productService: productService}
}

/*
func (h *productHandler) PostProduct(c *gin.Context) {
	var postProductReq dto.PostProductReq

	if err := c.ShouldBindJSON(&postProductReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid product data",
			"details": err.Error(),
		})
		return
	}

	var response dto.PostProductResp
	var err error
	if response.ID, err = h.productService.CreateProduct(context.Background(), postProductReq.ConvertToSvc()); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, response)
} */

func (h *productHandler) GetProduct(c *gin.Context) {
	id := c.Query("id")

	product, err := h.productService.GetProductByID(context.Background(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *productHandler) SelectProducts(c *gin.Context) {
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
		c.Error(err)
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

func (h *productHandler) SearchProductsByName(c *gin.Context) {
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
		c.Error(err)
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

func (h *productHandler) SelectFilteredProducts(c *gin.Context) {
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

	products, total, err := h.productService.SelectProductsByCategoryAndAttributes(c.Request.Context(), categoryID, filters, limit, offset)
	if err != nil {
		c.Error(err)
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


func (h *productHandler) GetStoreProducts(c *gin.Context) {
	id := c.Param("id")

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

	products, total, err := h.productService.GetStoreProducts(context.Background(), id, offset, limit)
	if err != nil {
		c.Error(err)
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

func (h *productHandler) PatchProduct(c *gin.Context) {
	id := c.Param("id")

	var update dto.PatchProductReq
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid update data",
			"details": err.Error(),
		})
		return
	}

	if err := h.productService.UpdateProduct(context.Background(), id, update.ConvertToSvc()); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}
