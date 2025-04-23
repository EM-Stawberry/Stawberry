package handler

import (
	"context"
	"fmt"
	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"

	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type OfferService interface {
	CreateOffer(ctx context.Context, offer offer.Offer) (uint, error)
	GetUserOffers(ctx context.Context, userID uint, limit, offset int) ([]entity.Offer, int64, error)
	GetOffer(ctx context.Context, offerID uint) (entity.Offer, error)
	UpdateOfferStatus(ctx context.Context, offerID uint, userID uint, status string) (entity.Offer, error)
	DeleteOffer(ctx context.Context, offerID uint) (entity.Offer, error)
}

type offerHandler struct {
	offerService OfferService
}

func NewOfferHandler(offerService OfferService) offerHandler {
	return offerHandler{offerService: offerService}
}

func (h *offerHandler) PostOffer(c *gin.Context) {
	userID, _ := c.Get("userID")

	var offer dto.PostOfferReq
	if err := c.ShouldBindJSON(&offer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offer.UserID = userID.(uint)
	offer.Status = "pending"
	offer.ExpiresAt = time.Now().Add(24 * time.Hour)

	var response dto.PostOfferResp
	var err error
	if response.ID, err = h.offerService.CreateOffer(context.Background(), offer.ConvertToSvc()); err != nil {
		c.Error(err)
		return
	}

	// Create notification for store
	// notification := models.Notification{
	// 	UserID:  offer.StoreID, // Store notification
	// 	OfferID: offer.ID,
	// 	Message: fmt.Sprintf("New offer received for product %d", offer.ProductID),
	// }
	// h.notifyRepo.Create(&notification)

	c.JSON(http.StatusCreated, offer)
}

func (h *offerHandler) GetUserOffers(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UserID"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
		return
	}

	offset := (page - 1) * limit

	offers, total, err := h.offerService.GetUserOffers(context.Background(), userID.(uint), offset, limit)
	if err != nil {
		c.Error(err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": offers,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	})
}

func (h *offerHandler) GetOffer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid non digit offer id"})
		return
	}

	offer, err := h.offerService.GetOffer(context.Background(), uint(id))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": offer,
	})
}

func (h *offerHandler) PatchOfferStatus(c *gin.Context) {
	// TODO: zap debug coverage
	ctx, ctxCancel := context.WithTimeout(c.Request.Context(), time.Second*10)
	defer ctxCancel()

	id, err := strconv.Atoi(c.Param("offerID"))
	if err != nil {
		c.Error(apperror.New(apperror.BadRequest, "offerID must be numeric", err))
		return
	}
	if id <= 0 {
		c.Error(apperror.New(apperror.BadRequest, "offerID must be positive", nil))
		return
	}

	var req dto.PatchOfferStatusReq
	if err = c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.New(apperror.BadRequest, "status field not provided", err))
		return
	}

	validStatuses := map[string]struct{}{
		"accepted":  {},
		"declined":  {},
		"cancelled": {}, // should be coming from a buyer
	}
	if _, ok := validStatuses[req.Status]; !ok {
		c.Error(apperror.New(apperror.BadRequest, "invalid status field value", nil))
		return
	}

	tmp, ok := c.Get("user")
	if !ok {
		c.Error(apperror.New(apperror.InternalError, "user context not found",
			fmt.Errorf("if we're here - someone changed the key at the bottom of auth middleware")))
		return
	}

	updatedOffer, err := h.offerService.UpdateOfferStatus(ctx, uint(id), tmp.(entity.User).ID, req.Status)
	if err != nil {
		c.Error(err)
		return
	}

	// TODO: notify the user about offer status change

	// Create notification for store
	// notification := models.Notification{
	// 	UserID:  offer.StoreID, // Store notification
	// 	OfferID: offer.ID,
	// 	Message: fmt.Sprintf("Offer %d has changed status to %s", offer.ID, offer.Status),
	// }
	// h.notifyRepo.Create(&notification)

	c.JSON(http.StatusOK, updatedOffer)
}

func (h *offerHandler) DeleteOffer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid nondigit offer id"})
		return
	}

	offer, err := h.offerService.DeleteOffer(context.Background(), uint(id))
	if err != nil {
		c.Error(err)
		return
	}

	// Create notification for store
	// notification := models.Notification{
	// 	UserID:  offer.StoreID, // Store notification
	// 	OfferID: offer.ID,
	// 	Message: fmt.Sprintf("Offer %d canceled", offer.ID),
	// }
	// h.notifyRepo.Create(&notification)

	c.JSON(http.StatusCreated, offer)
}
