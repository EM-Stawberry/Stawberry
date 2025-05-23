package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
	"github.com/Masterminds/squirrel"

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
	offerEntity entity.Offer,
	userID uint,
	isStore bool,
) (entity.Offer, error) {
	offer := model.ConvertOfferEntityToModel(offerEntity)

	updateOfferStatusQuery, args := squirrel.Update("offers").
		Set("status", offer.Status).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": offer.ID}).
		Suffix("returning status").
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return entity.Offer{}, apperror.New(apperror.DatabaseError, "failed to begin transaction", err)
	}
	defer tx.Rollback()

	err = isPendingOffer(ctx, offer.ID, tx)
	if err != nil {
		return entity.Offer{}, err
	}

	if isStore {
		err = isUserShopOwner(ctx, offer.ID, userID, tx)
		if err != nil {
			return entity.Offer{}, err
		}

	} else {
		// Если запрос на обновление статуса отправляет НЕ магазин, то добавляем проверку user_id
		// в квери, чтобы убедиться, что пользователь является создателем оффера.
		updateOfferStatusQuery, args = squirrel.Update("offers").
			Set("status", offer.Status).
			Set("updated_at", time.Now()).
			Where(squirrel.Eq{"id": offer.ID, "user_id": userID}).
			Suffix("returning status").
			PlaceholderFormat(squirrel.Dollar).
			MustSql()
	}

	err = tx.QueryRowx(updateOfferStatusQuery, args...).StructScan(&offer)
	if err != nil {
		return entity.Offer{}, apperror.New(apperror.DatabaseError, "error scanning into struct", err)
	}

	err = tx.Commit()
	if err != nil {
		return entity.Offer{}, apperror.New(apperror.DatabaseError, "failed to commit transaction", err)
	}

	return offer.ConvertToEntity(), nil
}

func isUserShopOwner(ctx context.Context, offerID, userID uint, tx *sqlx.Tx) error {
	validateShopOwnerIDQuery, args := squirrel.Select("users.id").
		From("users").
		InnerJoin("shops on users.id = shops.user_id").
		InnerJoin("offers on shops.id = offers.shop_id").
		Where(squirrel.Eq{"offers.id": offerID}).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	var requiredID uint
	err := tx.QueryRowContext(ctx, validateShopOwnerIDQuery, args...).Scan(&requiredID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrUserNotFound
		}
		return apperror.New(apperror.InternalError, "error scanning into uint", err)
	}

	if userID != requiredID {
		return apperror.New(apperror.Unauthorized, "unauthorized to update offer status", nil)
	}

	return nil
}

func isPendingOffer(ctx context.Context, offerID uint, tx *sqlx.Tx) error {
	getOfferStatusQuery, args := squirrel.Select("offers.status = 'pending'").
		From("offers").
		Where(squirrel.Eq{"offers.id": offerID}).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	var ok bool
	err := tx.QueryRowxContext(ctx, getOfferStatusQuery, args...).Scan(&ok)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrOfferNotFound
		}
		return apperror.New(apperror.InternalError, "error scanning offer status", err)
	}

	if !ok {
		return apperror.New(apperror.Conflict, "offer is not in a pending status", nil)
	}

	return nil
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
