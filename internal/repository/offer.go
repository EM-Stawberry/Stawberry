package repository

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type OfferRepository struct {
	db *sqlx.DB
}

func NewOfferRepository(db *sqlx.DB) *OfferRepository {
	return &OfferRepository{db: db}
}

func (r *OfferRepository) InsertOffer(
	ctx context.Context,
	offer offer.Offer,
) (uint, error) {

	_ = ctx
	_ = offer

	return offer.ID, nil
}

func (r *OfferRepository) GetOfferByID(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	var offer entity.Offer

	_ = ctx
	_ = offerID

	return offer, nil
}

func (r *OfferRepository) SelectUserOffers(
	ctx context.Context,
	userID uint,
	limit, offset int,
) ([]entity.Offer, int64, error) {
	var total int64

	var offers []entity.Offer

	_ = ctx
	_ = userID
	_ = limit
	_ = offset

	return offers, total, nil
}

func (r *OfferRepository) UpdateOfferStatus(
	ctx context.Context,
	offerID uint,
	status string,
) (entity.Offer, error) {

	var offer entity.Offer

	_ = ctx
	_ = offerID
	_ = status

	return offer, nil
}

func (r *OfferRepository) DeleteOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	var offer entity.Offer

	_ = ctx
	_ = offerID

	return offer, nil
}
