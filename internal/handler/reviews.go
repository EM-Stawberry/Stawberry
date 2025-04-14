package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
	"github.com/zuzaaa-dev/stawberry/internal/handler/dto"
	"go.uber.org/zap"
)

type ProductReviewService interface {
	AddReview(ctx context.Context, productID int, userID int, rating int, review string) error
	GetReviewsByProductID(ctx context.Context, productID int) ([]entity.ProductReview, error)
}

type productReviewHandler struct {
	prs    ProductReviewService
	logger *zap.Logger
}

func NewProductReviewHandler(prs ProductReviewService, l *zap.Logger) *productReviewHandler {
	return &productReviewHandler{
		prs: prs, logger: l,
	}
}

func (h *productReviewHandler) AddReview(c *gin.Context) {
	const op = "productReviewHandler.AddReviews()"
	log := h.logger.With(zap.String("op", op))

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid productID"})
		log.Warn("Failed to parse productID", zap.Error(err))
		return
	}

	var addReview = dto.AddReview
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

	err = h.prs.AddReview(c.Request.Context(), productID, uid, addReview.Rating, addReview.Review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add review"})
		log.Warn("Failed to add review", zap.Error(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Review added successfully"})
}

func (h *productReviewHandler) GetReviews(c *gin.Context) {
	const op = "productReviewHandler.GetReviews()"
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
