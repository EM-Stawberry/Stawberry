package offer

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/pkg/email"
)

type Repository interface {
	InsertOffer(ctx context.Context, offer Offer) (uint, error)
	GetOfferByID(ctx context.Context, offerID uint) (entity.Offer, error)
	SelectUserOffers(ctx context.Context, userID uint, limit, offset int) ([]entity.Offer, int64, error)
	UpdateOfferStatus(ctx context.Context, offer entity.Offer, userID uint, isStore bool) (entity.Offer, error)
	DeleteOffer(ctx context.Context, offerID uint) (entity.Offer, error)
}

const (
	statusAccepted  = "accepted"
	statusDeclined  = "declined"
	statusCancelled = "cancelled"
)

type Service struct {
	offerRepository Repository
	mailer          email.MailerService
}

func NewService(offerRepository Repository, mailer email.MailerService) *Service {
	return &Service{offerRepository: offerRepository, mailer: mailer}
}

func (os *Service) CreateOffer(
	ctx context.Context,
	offer Offer,
) (uint, error) {
	return os.offerRepository.InsertOffer(ctx, offer)
}

func (os *Service) GetOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	return os.offerRepository.GetOfferByID(ctx, offerID)
}

func (os *Service) GetUserOffers(
	ctx context.Context,
	userID uint,
	limit,
	offset int,
) ([]entity.Offer, int64, error) {
	return os.offerRepository.SelectUserOffers(ctx, userID, limit, offset)
}

func (os *Service) UpdateOfferStatus(
	ctx context.Context,
	offer entity.Offer,
	userID uint,
	isStore bool,
) (entity.Offer, error) {
	if isStore {
		validStatusesShop := map[string]struct{}{
			statusAccepted: {},
			statusDeclined: {},
		}
		if _, ok := validStatusesShop[offer.Status]; !ok {
			return entity.Offer{}, apperror.New(apperror.BadRequest, "invalid status field value", nil)
		}

	} else {
		validStatusesBuyer := map[string]struct{}{
			statusCancelled: {},
		}
		if _, ok := validStatusesBuyer[offer.Status]; !ok {
			return entity.Offer{}, apperror.New(apperror.BadRequest, "invalid status field value", nil)
		}
	}

	offerResp, err := os.offerRepository.UpdateOfferStatus(ctx, offer, userID, isStore)

	return offerResp, err
}

func (os *Service) DeleteOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	return os.offerRepository.DeleteOffer(ctx, offerID)
}
