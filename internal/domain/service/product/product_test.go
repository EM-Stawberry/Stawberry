package product

import (
	"context"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product/mocks"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnrichProducts", func() {
	var (
		mockCtrl *gomock.Controller
		mockRepo *mocks.MockRepository
		svc      *Service
		ctx      context.Context
		products []entity.Product
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockRepo = mocks.NewMockRepository(mockCtrl)
		svc = &Service{ProductRepository: mockRepo}
		ctx = context.Background()

		products = []entity.Product{
			{ID: 1, Name: "Product A"},
			{ID: 2, Name: "Product B"},
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("successfully enriches products", func() {
		mockRepo.EXPECT().GetPriceRangeByProductID(ctx, 1).Return(100.0, 200.0, nil)
		mockRepo.EXPECT().GetAverageRatingByProductID(ctx, 1).Return(4.5, 10, nil)
		mockRepo.EXPECT().GetPriceRangeByProductID(ctx, 2).Return(150.0, 300.0, nil)
		mockRepo.EXPECT().GetAverageRatingByProductID(ctx, 2).Return(4.2, 5, nil)

		result, err := svc.enrichProducts(ctx, products)
		Expect(err).ToNot(HaveOccurred())
		Expect(result[0].MinimalPrice).To(Equal(100.0))
		Expect(result[0].MaximalPrice).To(Equal(200.0))
		Expect(result[0].AverageRating).To(Equal(4.5))
		Expect(result[0].CountReviews).To(Equal(10))
		Expect(result[1].MinimalPrice).To(Equal(150.0))
		Expect(result[1].MaximalPrice).To(Equal(300.0))
		Expect(result[1].AverageRating).To(Equal(4.2))
		Expect(result[1].CountReviews).To(Equal(5))
	})

	It("returns error if GetPriceRangeByProductID fails", func() {
		mockRepo.EXPECT().GetPriceRangeByProductID(ctx, 1).Return(0.0, 0.0, errors.New("db error"))

		_, err := svc.enrichProducts(ctx, products)
		Expect(err).To(MatchError("db error"))
	})

	It("returns error if GetAverageRatingByProductID fails", func() {
		mockRepo.EXPECT().GetPriceRangeByProductID(ctx, 1).Return(100.0, 200.0, nil)
		mockRepo.EXPECT().GetAverageRatingByProductID(ctx, 1).Return(0.0, 0, errors.New("rating error"))

		_, err := svc.enrichProducts(ctx, products)
		Expect(err).To(MatchError("rating error"))
	})
})
