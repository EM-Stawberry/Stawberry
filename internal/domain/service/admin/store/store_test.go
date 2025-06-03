package store_test

import (
	"context"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/admin/store"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/repository/admin/store/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("StoreService.CreateUser", func() {
	var (
		mockRepo *mocks.MockRepositoryStore
		service  *store.Store
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		mockRepo = mocks.NewMockRepositoryStore(ctrl)

		service = store.NewStoreService(mockRepo)
	})

	Context("when InsertStore succeeds", func() {
		It("returns nil", func() {
			user := user.User{
				Email: "store@mail.com",
			}

			mockRepo.EXPECT().
				InsertStore(gomock.Any(), user).
				Return(nil)

			err := service.CreateUser(context.Background(), user)
			Expect(err).To(BeNil())
		})
	})

	Context("when InsertStore fails", func() {
		It("returns the error", func() {
			user := user.User{
				Email: "store@mail.com",
			}

			mockErr := errors.New("insert failed")

			mockRepo.EXPECT().
				InsertStore(gomock.Any(), user).
				Return(mockErr)

			err := service.CreateUser(context.Background(), user)
			Expect(err).To(MatchError("insert failed"))
		})
	})
})
