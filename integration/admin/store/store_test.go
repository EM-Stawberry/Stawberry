package store_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/integration/suite"
	storeService "github.com/EM-Stawberry/Stawberry/internal/domain/service/admin/store"
	"github.com/EM-Stawberry/Stawberry/internal/handler/admin/dto"
	storeHandler "github.com/EM-Stawberry/Stawberry/internal/handler/admin/store"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	storeRepo "github.com/EM-Stawberry/Stawberry/internal/repository/admin/store"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("RepositoryStore.InsertStore", func() {
	var (
		tdb    *suite.TestDatabase
		router *gin.Engine
		ctx    context.Context
	)

	BeforeEach(func() {
		tdb = suite.NewTestDB(GinkgoTB(), suite.WithMigrations("../../../migrations"))
		ctx = context.Background()
		cfg := config.LoadConfig()
		repo := storeRepo.NewRepositoryStore(tdb.DB, zap.NewNop())
		service := storeService.NewStoreService(repo)
		handler := storeHandler.NewStoreHandler(cfg, service)

		router = gin.New()
		router.Use(middleware.Errors())
		router.POST("/admin/store", handler.Registration)
	})

	AfterEach(func() {
		tdb.Close(ctx)
	})

	It("registers store and returns 201", func() {
		payload := dto.RegisterStoreReq{
			Name:     "Test Store",
			Email:    "store@mail.com",
			Phone:    "1234567890",
			Password: "secure123",
			IsStore:  true,
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/admin/store", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusCreated))

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		Expect(err).To(BeNil())
		Expect(resp["message"]).To(ContainSubstring("Store registered successfully"))
	})

	It("returns 400 on invalid payload", func() {
		body := `{"email": "bad"}`

		req := httptest.NewRequest(http.MethodPost, "/admin/store", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})
})
