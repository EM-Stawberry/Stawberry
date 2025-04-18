package reviews

import (
	"context"
	"fmt"

	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
	"go.uber.org/zap"
)

type SellerReviewRepository interface {
	AddReview(ctx context.Context, sellerID int, userID int, rating int, review string) (int, error)
	GetReviewsBySellerID(ctx context.Context, sellerID int) ([]entity.SellerReview, error)
	GetSellerByID(ctx context.Context, sellerID int) error
}

type sellerReviewService struct {
	srs    SellerReviewRepository
	logger *zap.Logger
}

func NewSellerReviewService(srr SellerReviewRepository, l *zap.Logger) *sellerReviewService {
	return &sellerReviewService{
		srs:    srr,
		logger: l,
	}
}

func (s *sellerReviewService) AddReview(
	ctx context.Context, sellerID int, userID int, rating int, review string,
) (
	int, error,
) {
	const op = "sellerReviewService.AddReview()"
	log := s.logger.With(zap.String("op", op))

	log.Info("Existence check")
	err := s.srs.GetSellerByID(ctx, sellerID)
	if err != nil {
		log.Warn("Seller not found", zap.Int("sellerID", sellerID), zap.Error(err))
		return 0, fmt.Errorf("op: %s, err: %s", op, err.Error())
	}

	log.Info("Adding a review")
	sellerID, err = s.AddReview(ctx, sellerID, userID, rating, review)
	if err != nil {
		log.Warn("Failed to add review", zap.Error(err))
		return 0, fmt.Errorf("op: %s, err: %s", op, err.Error())
	}

	log.Info("Review added successfully")
	return sellerID, nil
}

func (s *sellerReviewService) GetReviewsBySellerID(
	ctx context.Context, sellerID int,
) (
	[]entity.SellerReview, error,
) {
	const op = "sellerReviewService.GetReviewsBySellerID()"
	log := s.logger.With(zap.String("op", op))

	log.Info("Existence check")
	err := s.srs.GetSellerByID(ctx, sellerID)
	if err != nil {
		log.Warn("Seller not found", zap.Int("sellerID", sellerID), zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %s", op, err.Error())
	}

	log.Info("Receiving reviews")
	reviews, err := s.GetReviewsBySellerID(ctx, sellerID)
	if err != nil {
		log.Warn("Failed to get reviews", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %s", op, err.Error())
	}

	log.Info("Reviews received successfully")
	return reviews, nil
}
