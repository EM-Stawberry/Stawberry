package handler

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"

	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type OfferService interface {
	CreateOffer(ctx context.Context, offer entity.Offer) (uint, error)
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

func (h *OfferHandler) RegisterRoutes(group gin.IRoutes) {
	group.PATCH("offers/:offerID", h.PatchOfferStatus)
	group.POST("offers", h.PostOffer)
}

// @summary Create offer
// @tags offer
// @accept json
// @produce json
// @param body body dto.PostOfferReq true "Offer creation request"
// @success 201 {object} dto.PostOfferResp
// @failure 400 {object} apperror.Error
// @failure 401 {object} apperror.Error
// @failure 403 {object} apperror.Error
// @failure 500 {object} apperror.Error
// @Router /offers [post]
func (h *OfferHandler) PostOffer(c *gin.Context) {
	ctx, ctxCancel := context.WithTimeout(c.Request.Context(), time.Second*30)
	defer ctxCancel()

	store, ok := helpers.UserIsStoreContext(c)
	if !ok {
		_ = c.Error(apperror.New(apperror.Unauthorized, "invalid credentials", nil))
		return
	}
	if store {
		_ = c.Error(apperror.New(apperror.Forbidden,
			"store accounts are not allowed to create offer", nil))
		return
	}

	var offerPost dto.PostOfferReq
	if err := c.ShouldBindJSON(&offerPost); err != nil {
		_ = c.Error(apperror.New(apperror.BadRequest, "Invalid offer data", err))
		return
	}

	userID, ok := helpers.UserIDContext(c)
	if !ok {
		_ = c.Error(apperror.New(apperror.Unauthorized, "invalid credentials", nil))
		return
	}

	offerEnt := offerPost.ConvertToEntity()
	offerEnt.UserID = userID

	offerID, err := h.offerService.CreateOffer(ctx, offerEnt)
	if err != nil {
		_ = c.Error(apperror.New(apperror.InternalError, "Failed to create offer", err))
		return
	}

	c.JSON(http.StatusCreated, dto.PostOfferResp{ID: offerID})
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
	ctx, ctxCancel := context.WithTimeout(c.Request.Context(), time.Second*30)
	defer ctxCancel()

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

	iisStore, ok := c.Get(helpers.UserIsStoreKey)
	if !ok {
		_ = c.Error(apperror.New(apperror.InternalError,
			"user isstore key not found in ctx", nil))
	}
	usrIsStore := iisStore.(bool)

	offerEntity := req.ConvertToEntity()
	offerEntity.ID = uint(id)

	updatedOffer, err := h.offerService.UpdateOfferStatus(ctx, offerEntity, usrID, usrIsStore)
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
