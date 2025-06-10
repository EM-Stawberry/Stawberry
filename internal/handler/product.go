package handler

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"
	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"go.uber.org/zap"

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
	logger         *zap.Logger
}

func NewProductHandler(productService ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		logger:         logger,
	}
}

// PostProduct godoc
// @Summary Создать новый продукт
// @Description Создает новый продукт в системе
// @Tags products
// @Accept json
// @Produce json
// @Param productData body dto.PostProductReq true "Данные продукта"
// @Success 201 {object} dto.PostProductResp
// @Failure 400 {object} object "Неверные данные продукта"
// @Failure 500 {object} object "Ошибка создания"
// @Router /products/ [post]
func (h *ProductHandler) PostProduct(c *gin.Context) {
	var postProductReq dto.PostProductReq

	if err := c.ShouldBindJSON(&postProductReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product data",
			"details": err.Error(),
		})
		h.logger.Error("Invalid product data, PostProduct, productHandler", zap.Error(err))
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
		h.logger.Error("Failed to create product, PostProduct, productHandler", zap.Error(err))
		return
	}

	c.JSON(http.StatusCreated, dto.PostProductResp{ID: id})
}

// GetProduct godoc
// @Summary Получить продукт по ID
// @Description Возвращает продукт по указанному ID
// @Tags products
// @Produce json
// @Param product_id path int true "ID продукта"
// @Success 200 {object} dto.GetProductResp
// @Failure 400 {object} object "Неверный ID продукта"
// @Failure 500 {object} object "Ошибка сервера"
// @Router /products/{product_id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product ID",
			"details": err.Error(),
		})
		h.logger.Error("Invalid product ID, GetProduct, productHandler", zap.Error(err))
		return
	}

	product, err := h.productService.GetProductByID(context.TODO(), (uint)(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get product",
			"details": err.Error(),
		})
		h.logger.Error("Failed to get product, GetProduct, productHandler", zap.Error(err))
		return
	}

	response := &dto.GetProductResp{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		CategoryID:  product.CategoryID,
	}

	c.JSON(http.StatusOK, response)
}

// GetProducts godoc
// @Summary Получить список продуктов
// @Description Возвращает список продуктов с пагинацией
// @Tags products
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Success 200 {object} object{data=[]dto.GetProductResp,meta=object{current_page=int,
// per_page=int,total_items=int,total_pages=int}}
// @Failure 400 {object} object "Неверные параметры запроса"
// @Failure 500 {object} object "Ошибка сервера"
// @Router /products/ [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid page number",
			"details": err.Error(),
		})
		h.logger.Error("Invalid page number, GetProducts, productHandler", zap.Error(err))
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid limit value (should be between 1 and 100)",
			"details": err.Error(),
		})
		h.logger.Error("Invalid limit value, GetProducts, productHandler", zap.Error(err))
		return
	}

	offset := (page - 1) * limit

	products, total, err := h.productService.GetProducts(context.TODO(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get products",
			"details": err.Error(),
		})
		h.logger.Error("Failed to get products, GetProducts, productHandler", zap.Error(err))
		return
	}

	responce := make([]*dto.GetProductResp, 0, len(products))

	for _, product := range products {
		getProductResp := &dto.GetProductResp{
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

// GetStoreProducts godoc
// @Summary Получить продукты магазина
// @Description Возвращает список продуктов для указанного магазина
// @Tags products
// @Produce json
// @Param shop_id path int true "ID магазина"
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Success 200 {object} object{data=[]dto.GetProductResp,meta=object{current_page=int,
// per_page=int,total_items=int,total_pages=int}}
// @Failure 400 {object} object "Неверные параметры запроса"
// @Failure 500 {object} object "Ошибка сервера"
// @Router /products/shop/{shop_id} [get]
func (h *ProductHandler) GetStoreProducts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("store_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product ID",
			"details": err.Error(),
		})
		h.logger.Error("Invalid product ID, GetStoreProducts, productHandler", zap.Error(err))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid page number",
			"details": err.Error(),
		})
		h.logger.Error("Invalid page number, GetStoreProducts, productHandler", zap.Error(err))
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid limit value (should be between 1 and 100)",
			"details": err.Error(),
		})
		h.logger.Error("Invalid limit value, GetStoreProducts, productHandler", zap.Error(err))
		return
	}

	offset := (page - 1) * limit

	products, total, err := h.productService.GetStoreProducts(context.Background(), (uint)(id), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get products",
			"details": err.Error(),
		})
		h.logger.Error("Failed to get products, GetStoreProducts, productHandler", zap.Error(err))
		return
	}

	responce := make([]*dto.GetProductResp, 0, len(products))

	for _, product := range products {
		getProductResp := &dto.GetProductResp{
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

// PatchProduct godoc
// @Summary Обновить продукт
// @Description Обновляет информацию о продукте
// @Tags products
// @Accept json
// @Produce json
// @Param product_id path int true "ID продукта"
// @Param updateData body dto.PatchProductReq true "Данные для обновления"
// @Success 200 {object} object "Продукт обновлен"
// @Failure 400 {object} object "Неверные данные"
// @Failure 500 {object} object "Ошибка обновления"
// @Router /products/{product_id} [patch]
func (h *ProductHandler) PatchProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product ID",
			"details": err.Error(),
		})
		h.logger.Error("Invalid product ID, PatchProduct, productHandler", zap.Error(err))
		return
	}

	var updateReq dto.PatchProductReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid update data",
			"details": err.Error(),
		})
		h.logger.Error("Invalid update data, PatchProduct, productHandler", zap.Error(err))
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
		h.logger.Error("Failed to update product, PatchProduct, productHandler", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

// DeleteProduct godoc
// @Summary Удалить продукт
// @Description Удаляет продукт по указанному ID
// @Tags products
// @Produce json
// @Param product_id path int true "ID продукта"
// @Success 200 {object} object "Продукт удален"
// @Failure 400 {object} object "Неверный ID продукта"
// @Failure 500 {object} object "Ошибка удаления"
// @Router /products/{product_id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid product ID",
			"details": err.Error(),
		})
		h.logger.Error("Invalid product ID, DeleteProduct, productHandler", zap.Error(err))
		return
	}

	err = h.productService.DeleteProduct(context.TODO(), (uint)(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to delete product",
			"details": err.Error(),
		})
		h.logger.Error("Failed to delete product, DeleteProduct, productHandler", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
