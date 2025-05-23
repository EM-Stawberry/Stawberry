package offer

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type Repository interface {
	InsertOffer(ctx context.Context, offer Offer) (uint, error)
	GetOfferByID(ctx context.Context, offerID uint) (entity.Offer, error)
	SelectUserOffers(ctx context.Context, userID uint, limit, offset int) ([]entity.Offer, int64, error)
	UpdateOfferStatus(ctx context.Context, offerID uint, status string) (entity.Offer, error)
	DeleteOffer(ctx context.Context, offerID uint) (entity.Offer, error)
}

type Service struct {
	offerRepository Repository
}

func NewService(offerRepository Repository) *Service {
	return &Service{offerRepository: offerRepository}
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
	offerID uint,
	status string,
) (entity.Offer, error) {
	return os.offerRepository.UpdateOfferStatus(ctx, offerID, status)
}

func (os *Service) DeleteOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	return os.offerRepository.DeleteOffer(ctx, offerID)
}
