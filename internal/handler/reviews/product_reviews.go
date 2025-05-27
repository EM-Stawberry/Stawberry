package reviews

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
	"github.com/zuzaaa-dev/stawberry/internal/handler/reviews/dto"
	"go.uber.org/zap"
)

type ProductReviewsService interface {
	AddReview(ctx context.Context, productID int, userID int, rating int, review string) (int, error)
	GetReviewsByProductID(ctx context.Context, productID int) ([]entity.ProductReview, error)
}

type ProductReviewsHandler struct {
	prs    ProductReviewsService
	logger *zap.Logger
}

func NewProductReviewHandler(prs ProductReviewsService, l *zap.Logger) *ProductReviewsHandler {
	return &ProductReviewsHandler{
		prs:    prs,
		logger: l,
	}
}

// AddReview godoc
// @Summary Добавление отзыва о продукте
// @Description Добавляет новый отзыв о продукте
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param review body dto.AddReviewDTO true "Данные отзыва"
// @Security BearerAuth
// @Success 201 {object} map[string]string "Отзыв успешно добавлен"
// @Failure 400 {object} map[string]string "Некорректный ввод"
// @Failure 401 {object} map[string]string "Неавторизованный доступ"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/products/{id}/reviews [post]
func (h *ProductReviewsHandler) AddReview(c *gin.Context) {
	const op = "productReviewsHandler.AddReviews()"
	log := h.logger.With(zap.String("op", op))

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid productID"})
		log.Warn("Failed to parse productID", zap.Error(err))
		return
	}

	var addReview dto.AddReviewDTO
	if err := c.ShouldBindJSON(&addReview); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		log.Warn("Failed to bind JSON", zap.Error(err))
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		log.Warn("Failed to get userID from context", zap.Error(err))
		return
	}

	uid, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid userID type"})
		log.Warn("Invalid userID type")
		return
	}

	id, err := h.prs.AddReview(c.Request.Context(), productID, uid, addReview.Rating, addReview.Review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add review"})
		log.Warn("Failed to add review", zap.Int("id: ", id), zap.Error(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "review added successfully"})
}

// GetReviews godoc
// @Summary Получение списка отзывов о продукте
// @Description Получает все отзывы о продукте по его ID
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {array} entity.ProductReview "Список отзывов"
// @Failure 400 {object} map[string]string "Некорректный ID продукта"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/products/{id}/reviews [get]
func (h *ProductReviewsHandler) GetReviews(c *gin.Context) {
	const op = "productReviewsHandler.GetReviews()"
	log := h.logger.With(zap.String("op", op))

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid productID"})
		log.Warn("Failed to parse productID", zap.Error(err))
		return
	}

	reviews, err := h.prs.GetReviewsByProductID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reviews"})
		log.Warn("Failed to fetch reviews", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, reviews)
}
