package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
	"go.uber.org/zap"
)

type sellerReviewsRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewSellerReviewRepository(db *sqlx.DB, l *zap.Logger) *sellerReviewsRepository {
	return &sellerReviewsRepository{
		db:     db,
		logger: l,
	}
}

func (r *sellerReviewsRepository) AddReview(
	ctx context.Context, sellerID int, userID int, rating int, review string,
) error {
	const op = "sellerReviewsRepository.AddReview()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.
		Insert("seller_reviews").
		Columns("seller_id", "user_id", "rating", "review").
		Values(sellerID, userID, rating, review).ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("op: %s, err: %s", op, err.Error())
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return fmt.Errorf("op: %s, err: %s", op, err.Error())
	}

	return nil
}

// GetReviewsBySellerID получает отзывы по ID продавца
func (r *sellerReviewsRepository) GetReviewsBySellerID(
	ctx context.Context, sellerID int,
) (
	[]entity.SellerReview, error,
) {
	const op = "sellerReviewsRepository.GetReviewsBySellerID()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.
		Select("id", "seller_id", "user_id", "rating", "review", "created_at").
		From("seller_reviews").
		Where("seller_id = ?", sellerID).ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %s", op, err)
	}

	var reviews []entity.SellerReview
	err = r.db.SelectContext(ctx, &reviews, query, args...)
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %s", op, err)
	}

	return reviews, nil
}

func (r *sellerReviewsRepository) GetSellerByID(
	ctx context.Context, sellerID int,
) (
	entity.SellerReview, error,
) {
	const op = "sellerReviewsRepository.GetSellerByID()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.
		Select("id", "name", "description").
		From("sellers").
		Where("id = ?", sellerID).ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return entity.SellerReview{}, fmt.Errorf("op: %s, err: %s", op, err)
	}

	var seller entity.SellerReview
	err = r.db.GetContext(ctx, &seller, query, args...)
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return entity.SellerReview{}, fmt.Errorf("op: %s, err: %s", op, err)
	}

	return seller, nil
}
