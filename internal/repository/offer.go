package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/Masterminds/squirrel"
	"log/slog"
	"time"

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

	// TODO: zap debug coverage

	var offer entity.Offer

	updateOfferStatusQuery, args := squirrel.Update("offers").
		Set("status", status).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": offerID, "status": "pending"}).
		Suffix("returning id, offer_price, status, created_at, " +
			"updated_at, user_id, product_id, shop_id").
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return offer, apperror.New(apperror.DatabaseError, "failed to begin transaction", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowx(updateOfferStatusQuery, args...).StructScan(&offer)
	if err != nil {
		return offer, apperror.New(apperror.DatabaseError, "error scanning into struct", err)
	}

	err = tx.Commit()
	if err != nil {
		return offer, apperror.New(apperror.DatabaseError, "failed to commit transaction", err)
	}

	return offer, nil
}

func (r *offerRepository) IsUserShopOwner(ctx context.Context, offerID, userID uint) (bool, error) {
	var requiredID uint

	validateShopOwnerIDQuery, args := squirrel.Select("users.id").
		From("users").
		InnerJoin("shops on users.id = shops.user_id").
		InnerJoin("offers on shops.id = offers.shop_id").
		Where(squirrel.Eq{"offers.id": offerID}).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	err := r.db.QueryRowContext(ctx, validateShopOwnerIDQuery, args...).Scan(&requiredID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, apperror.ErrOfferNotFound
		}
		return false, apperror.New(apperror.InternalError, "error scanning into uint", err)
	}

	if userID != requiredID {
		return false, apperror.New(apperror.Unauthorized, "unauthorized to update offer status", nil)
	}
	return true, nil
}

func (r *offerRepository) IsPendingOffer(ctx context.Context, offerID uint) (bool, error) {
	getOfferStatusQuery, args := squirrel.Select("offers.status = 'pending'").
		From("offers").
		Where(squirrel.Eq{"offers.id": offerID}).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	var ok bool
	err := r.db.QueryRowxContext(ctx, getOfferStatusQuery, args...).Scan(&ok)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, apperror.ErrOfferNotFound
		}
		slog.Error(err.Error())
		return false, apperror.New(apperror.InternalError, "error scanning offer status", err)
	}

	return ok, nil
}

func (r *offerRepository) DeleteOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	var offer entity.Offer

	return offer, nil
}
