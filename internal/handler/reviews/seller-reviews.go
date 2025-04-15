package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
	"go.uber.org/zap"
)

type SellerReviewsService interface {
	AddReviews(ctx context.Context, sellerID int, userID int, rating int, review string) error
	GetReviewsById(ctx context.Context, sellerID int) ([]entity.SellerReview, error)
}

type sellerReviewsHandler struct {
	srs    SellerReviewsService
	logger *zap.Logger
}

func NewSellerReviewsHandler(srs SellerReviewsService, l *zap.Logger) *sellerReviewsHandler {
	return &sellerReviewsHandler{
		srs:    srs,
		logger: l,
	}
}

func (h *sellerReviewsHandler) AddReviews(c *gin.Context) {
	const op = "sellerReviewsHandler.AddReviews()"
	log := h.logger.With(zap.String("op", op))

	sellerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sellerID"})
		log.Warn("Failed to parse productID", zap.Error(err))
		return
	}

	var addReview = AddReviewDTO
	if err := c.ShouldBindJSON(&addReview); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		log.Warn("Failed to bind JSON", zap.Error(err))
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not authenticated"})
		log.Warn("Failed to get userID from context", zap.Error(err))
		return
	}

	uid, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid userID type"})
		log.Warn("Invalid userID type")
		return
	}

	err = h.srs.AddReviews(c.Request.Context(), sellerID, uid, addReview.Rating, addReview.Review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add review"})
		log.Warn("Failed to add review", zap.Error(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "review added successfully"})
}

func (h *sellerReviewsHandler) GetReviews(c *gin.Context) {
	const op = "sellerReviewsHandler.GetReviews()"
	log := h.logger.With(zap.String("op", op))

	sellerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sellerID"})
		log.Warn("Failed to parse sellerID", zap.Error(err))
		return
	}

	reviews, err := h.srs.GetReviewsById(c.Request.Context(), sellerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reviews"})
		log.Warn("Failed to fetch reviews", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, reviews)
}
