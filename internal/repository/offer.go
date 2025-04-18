package repository

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type offerRepository struct {
	db *sqlx.DB
}

func NewOfferRepository(db *sqlx.DB) *offerRepository {
	return &offerRepository{db: db}
}

func (r *offerRepository) InsertOffer(
	ctx context.Context,
	offer offer.Offer,
) (uint, error) {

	return offer.ID, nil
}

func (r *offerRepository) GetOfferByID(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	var offer entity.Offer

	return offer, nil
}

func (r *offerRepository) SelectUserOffers(
	ctx context.Context,
	userID uint,
	limit, offset int,
) ([]entity.Offer, int64, error) {
	var total int64

	var offers []entity.Offer

	return offers, total, nil
}

func (r *offerRepository) UpdateOfferStatus(
	ctx context.Context,
	offerID uint,
	status string,
) (entity.Offer, error) {

	var offer entity.Offer

	return offer, nil
}

func (r *offerRepository) DeleteOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	var offer entity.Offer

	return offer, nil
}
