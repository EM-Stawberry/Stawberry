package handler

import (
	"context"
	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"
	"math"
	"net/http"
	"strconv"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"

	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type OfferService interface {
	CreateOffer(ctx context.Context, offer offer.Offer) (uint, error)
	GetUserOffers(ctx context.Context, userID uint, limit, offset int) ([]entity.Offer, int64, error)
	GetOffer(ctx context.Context, offerID uint) (entity.Offer, error)
	UpdateOfferStatus(ctx context.Context, offer entity.Offer, userID uint, isStore bool) (entity.Offer, error)
	DeleteOffer(ctx context.Context, offerID uint) (entity.Offer, error)
}

type OfferHandler struct {
	offerService OfferService
}

func NewOfferHandler(offerService OfferService) *OfferHandler {
	return &OfferHandler{offerService: offerService}
}

func (h *OfferHandler) PostOffer(c *gin.Context) {
	userID, _ := c.Get("userID")

	var offer dto.PostOfferReq
	if err := c.ShouldBindJSON(&offer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offer.UserID = userID.(uint)

	//var response dto.PostOfferResp
	//var err error
	//if response.ID, err = h.offerService.CreateOffer(context.Background(), offer.ConvertToEntity()); err != nil {
	//	_ = c.Error(err)
	//	return
	//}

	// Create notification for store
	// notification := models.Notification{
	// 	UserID:  offer.StoreID, // Store notification
	// 	OfferID: offer.ID,
	// 	Message: fmt.Sprintf("New offer received for product %d", offer.ProductID),
	// }
	// h.notifyRepo.Create(&notification)

	c.JSON(http.StatusCreated, offer)
}

func (h *OfferHandler) GetUserOffers(c *gin.Context) {
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
		_ = c.Error(err)
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

func (h *OfferHandler) GetOffer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid non digit offer id"})
		return
	}

	offerEntity, err := h.offerService.GetOffer(context.Background(), uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": offerEntity,
	})
}

// @summary Update offer status
// @tags offer
// @accept json
// @produce json
// @param id path int true "Offer ID"
// @param body body dto.PatchOfferStatusReq true "Offer status update request"
// @success 200 {object} dto.PatchOfferStatusResp
// @failure 400 {object} apperror.Error
// @failure 401 {object} apperror.Error
// @failure 404 {object} apperror.Error
// @failure 409 {object} apperror.Error
// @failure 500 {object} apperror.Error
// @Router /offers/{offerID} [patch]
func (h *OfferHandler) PatchOfferStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("offerID"))
	if err != nil {
		_ = c.Error(apperror.New(apperror.BadRequest, "offerID must be numeric", err))
		return
	}
	if id <= 0 {
		_ = c.Error(apperror.New(apperror.BadRequest, "offerID must be positive", nil))
		return
	}

	var req dto.PatchOfferStatusReq
	if err = c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.New(apperror.BadRequest, "status field not provided", err))
		return
	}

	usrID, ok := helpers.UserIDContext(c)
	if !ok {
		_ = c.Error(apperror.New(apperror.InternalError,
			"user id key not found in ctx", nil))
	}

	usrIsStore, ok := helpers.UserIsStoreContext(c)
	if !ok {
		_ = c.Error(apperror.New(apperror.InternalError,
			"user isstore key not found in ctx", nil))
	}

	offerEntity := req.ConvertToEntity()
	offerEntity.ID = uint(id)

	updatedOffer, err := h.offerService.UpdateOfferStatus(c.Request.Context(), offerEntity, usrID, usrIsStore)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.PatchOfferStatusResp{NewStatus: updatedOffer.Status})
}

func (h *OfferHandler) DeleteOffer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid nondigit offer id"})
		return
	}

	offer, err := h.offerService.DeleteOffer(context.Background(), uint(id))
	if err != nil {
		_ = c.Error(err)
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
