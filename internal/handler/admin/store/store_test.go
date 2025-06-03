package store_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/admin/store/mocks"
	"github.com/EM-Stawberry/Stawberry/internal/handler/admin/store"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var validStoreBody = `{
  "name": "Test Store",
  "email": "store@mail.com",
  "password": "12345678",
  "phone": "1234567890",
  "is_store": true
}`

var _ = Describe("HandlerStore.Registration", func() {
	var (
		router      *gin.Engine
		mockService *mocks.MockServiceStore
		handler     store.HandlerStore
		resp        map[string]string
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		mockCtrl := gomock.NewController(GinkgoT())
		mockService = mocks.NewMockServiceStore(mockCtrl)

		handler = store.NewStoreHandler(
			&config.Config{
				Token: config.TokenConfig{
					RefreshTokenDuration: 10,
				}}, mockService)

		router = gin.New()
		router.Use(middleware.Errors())
		router.POST("/admin/store", handler.Registration)
	})

	Context("valid input", func() {
		It("returns 201 Created", func() {
			mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodPost, "/admin/store", strings.NewReader(validStoreBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			err := json.Unmarshal(w.Body.Bytes(), &resp)
			Expect(err).To(BeNil(), fmt.Sprintf("Body: %s", w.Body.String()))
			Expect(resp["message"]).To(Equal("Store registered successfully"))
		})
	})

	Context("invalid JSON input", func() {
		It("returns 400 Bad Request", func() {
			body := `{"email": "invalid"}`

			req := httptest.NewRequest(http.MethodPost, "/admin/store", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("service returns error", func() {
		It("returns error from service", func() {
			mockService.
				EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Return(apperror.New(apperror.InternalError, "something failed", nil))

			req := httptest.NewRequest(http.MethodPost, "/admin/store", strings.NewReader(validStoreBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			fmt.Println("Status:", w.Code)
			fmt.Println("Body:", w.Body.String())
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
