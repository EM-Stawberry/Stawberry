package product

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"github.com/zuzaaa-dev/stawberry/internal/domain/service/product"

	"github.com/gin-gonic/gin"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *product.Product) (uint, error)
	GetProductByID(ctx context.Context, id uint) (*product.Product, error)
	GetProducts(ctx context.Context, offset, limit int) ([]*product.Product, int, error)
	GetStoreProducts(ctx context.Context, id uint, offset, limit int) ([]*product.Product, int, error)
	UpdateProduct(ctx context.Context, id uint, updateProduct *product.UpdateProduct) error
	DeleteProduct(ctx context.Context, id uint) error
}

type ProductHandler struct {
	productService ProductService
}

func NewProductHandler(productService ProductService) ProductHandler {
	return ProductHandler{productService: productService}
}

func (h *ProductHandler) PostProduct(c *gin.Context) {
	var postProductReq PostProductReq

	if err := c.ShouldBindJSON(&postProductReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product data",
			"details": err.Error(),
		})
		return
	}

	product := &product.Product{
		Name:        postProductReq.Name,
		Description: postProductReq.Description,
		CategoryID:  postProductReq.CategoryID,
		ShopPointID: postProductReq.ShopID,
		Price:       postProductReq.Price,
		Quantity:    postProductReq.Quantity,
	}

	id, err := h.productService.CreateProduct(context.TODO(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create product",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, PostProductResp{ID: id})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product ID",
			"details": err.Error(),
		})
		return
	}

	product, err := h.productService.GetProductByID(context.TODO(), (uint)(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get product",
			"details": err.Error(),
		})
		return
	}

	response := &GetProductResp{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		CategoryID:  product.CategoryID,
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid page number",
			"details": err.Error(),
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid limit value (should be between 1 and 100)",
			"details": err.Error(),
		})
		return
	}

	offset := (page - 1) * limit

	products, total, err := h.productService.GetProducts(context.TODO(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get products",
			"details": err.Error(),
		})
		return
	}

	responce := make([]*GetProductResp, 0, len(products))

	for _, product := range products {
		getProductResp := &GetProductResp{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			CategoryID:  product.CategoryID,
		}

		responce = append(responce, getProductResp)

	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": responce,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	})
}

func (h *ProductHandler) GetStoreProducts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("store_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product ID",
			"details": err.Error(),
		})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid page number",
			"details": err.Error(),
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid limit value (should be between 1 and 100)",
			"details": err.Error(),
		})
		return
	}

	offset := (page - 1) * limit

	products, total, err := h.productService.GetStoreProducts(context.Background(), (uint)(id), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get products",
			"details": err.Error(),
		})
		return
	}

	responce := make([]*GetProductResp, 0, len(products))

	for _, product := range products {
		getProductResp := &GetProductResp{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			CategoryID:  product.CategoryID,
		}

		responce = append(responce, getProductResp)

	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": responce,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	})
}

func (h *ProductHandler) PatchProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product ID",
			"details": err.Error(),
		})
		return
	}

	var updateReq PatchProductReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid update data",
			"details": err.Error(),
		})
		return
	}

	update := &product.UpdateProduct{
		Name:        updateReq.Name,
		Description: updateReq.Description,
		CategoryID:  updateReq.CategoryID,
		ShopPointID: updateReq.ShopPointID,
		Price:       updateReq.Price,
		Quantity:    updateReq.Quantity,
	}

	err = h.productService.UpdateProduct(context.TODO(), (uint)(id), update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update product",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product ID",
			"details": err.Error(),
		})
		return
	}

	err = h.productService.DeleteProduct(context.TODO(), (uint)(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to delete product",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
