package handler

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"

	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type OfferService interface {
	CreateOffer(ctx context.Context, offer offer.Offer) (uint, error)
	GetUserOffers(ctx context.Context, userID uint, page, limit int) ([]entity.Offer, int, error)
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

// @summary Get user's offers
// @tags offer
// @accept json
// @produce json
// @param page query int false "Page number for pagination" default(1)
// @param limit query int false "Number of items per page (5-100)" default(10)
// @success 200 {object} dto.GetUserOffersResp
// @failure 400 {object} apperror.Error
// @failure 500 {object} apperror.Error
// @Router /offers/user [get]
func (h *OfferHandler) GetUserOffers(c *gin.Context) {
	ctx, canc := context.WithTimeout(c.Request.Context(), time.Second*30)
	defer canc()

	userID, ok := helpers.UserIDContext(c)
	if !ok {
		_ = c.Error(apperror.New(apperror.InternalError, "user ID not found in context", nil))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		_ = c.Error(apperror.New(apperror.BadRequest, "invalid page number", err))
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 5 || limit > 100 {
		_ = c.Error(apperror.New(apperror.BadRequest, "invalid limit value (must be 5-100)", err))
		return
	}

	offersEnt, total, err := h.offerService.GetUserOffers(ctx, userID, page, limit)
	if err != nil {
		_ = c.Error(apperror.New(apperror.InternalError,
			fmt.Sprintf("failed to get user (userID: %d) offers", userID), err))
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	offersResp := dto.FormUserOffers(offersEnt, page, limit, total, totalPages)

	c.JSON(http.StatusOK, offersResp)
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
	ctx, ctxCancel := context.WithTimeout(c.Request.Context(), time.Second*10)
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

	//tmp, ok := c.Get("user")
	//usr := tmp.(entity.User)
	//if !ok {
	//	c.Error(apperror.New(apperror.InternalError, "user context not found",
	//		fmt.Errorf("if we're here - someone changed the key at the bottom of auth middleware")))
	//	return
	//}

	iid, ok := c.Get(helpers.UserIDKey)
	if !ok {
		_ = c.Error(apperror.New(apperror.InternalError,
			"user id key not found in ctx", nil))
	}
	usrID := iid.(uint)

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

	// Create notification for store
	// notification := models.Notification{
	// 	UserID:  offer.StoreID, // Store notification
	// 	OfferID: offer.ID,
	// 	Message: fmt.Sprintf("Offer %d has changed status to %s", offer.ID, offer.Status),
	// }
	// h.notifyRepo.Create(&notification)

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
