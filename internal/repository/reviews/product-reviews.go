package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
	"go.uber.org/zap"
)

type productReviewRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewProductReviewRepository(db *sqlx.DB, l *zap.Logger) *productReviewRepository {
	return &productReviewRepository{
		db:     db,
		logger: l,
	}
}

func (r *productReviewRepository) AddReview(
	ctx context.Context, productID int, userID int, rating int, review string,
) error {

	const op = "productReviewHandler.GetReviewsByProductID()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.
		Insert("product_reviews").
		Columns("product_id", "user_id", "rating", "review").
		Values(productID, userID, rating, review).ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("op: %s, err: %s", op, err.Error())
	}

	var reviews []entity.ProductReview
	err = r.db.SelectContext(ctx, &reviews, query, args...)
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return fmt.Errorf("op: %s, err: %s", op, err.Error())
	}

	return nil
}

func (r *productReviewRepository) GetProductByID(
	ctx context.Context, productID int,
) (
	entity.Product, error,
) {
	const op = "productReviewHandler.GetProductByID()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.
		Select("id", "product_id", "user_id", "rating", "review", "created_at").
		From("products").
		Where("id = ?", productID).ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return entity.Product{}, fmt.Errorf("op: %s, err: %s", op, err)
	}

	var product entity.Product
	err = r.db.GetContext(ctx, &product, query, args...)
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return entity.Product{}, fmt.Errorf("op: %s, err: %s", op, err)
	}

	return product, nil
}
