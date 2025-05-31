package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/EM-Stawberry/Stawberry/internal/handler"
	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	imageName  = "postgres:17.4-alpine"
	dbName     = "db_test"
	dbUser     = "postgres"
	dbPassword = "postgres"
)

func GetContainer() *postgres.PostgresContainer {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, imageName,
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(wait.ForLog(`database system is ready to accept connections`).
			WithOccurrence(2).WithPollInterval(time.Second)),
	)
	if err != nil {
		slog.Error("error starting container", "err", err.Error())
		return nil
	}

	err = pgContainer.Snapshot(context.Background())
	if err != nil {
		slog.Error("error snapshotting container", "err", err)
		return nil
	}

	return pgContainer
}

func GetDB(pgContainer *postgres.PostgresContainer) (*sqlx.DB, error) {
	connString, err := pgContainer.ConnectionString(context.Background(), "sslmode=disable")
	if err != nil {
		slog.Error(err.Error())
		_ = pgContainer.Terminate(context.Background())
		return nil, err
	}

	db, err := sqlx.Connect("pgx", connString)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		return nil, err
	}

	_ = goose.SetDialect("postgres")

	err = goose.Up(db.DB, `../../migrations`)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		slog.Error(err.Error())
		return nil, err
	}

	_, err = sqlx.LoadFile(db, `../testdata/offer/sql/populate_test_db.sql`)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return db, nil
}

func mockAuthShopOwnerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		mockUser := entity.User{
			ID:       1,
			Name:     "user1",
			Password: "no",
			Email:    "user1email",
			Phone:    "user1phone",
			IsStore:  true,
		}
		c.Set("user", mockUser)
		c.Set(helpers.UserIDKey, uint(1))
		c.Set(helpers.UserIsStoreKey, true)
		c.Next()
	}
}

func mockAuthBuyerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		mockUser := entity.User{
			ID:       2,
			Name:     "user2",
			Password: "no",
			Email:    "user2email",
			Phone:    "user2phone",
			IsStore:  false,
		}
		c.Set("user", mockUser)
		c.Set(helpers.UserIDKey, uint(2))
		c.Set(helpers.UserIsStoreKey, false)
		c.Next()
	}
}

func mockAuthIncorrectShopOwnerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		mockUser := entity.User{
			ID:       3,
			Name:     "user3",
			Password: "no",
			Email:    "user3email",
			Phone:    "user3phone",
			IsStore:  true,
		}
		c.Set("user", mockUser)
		c.Set(helpers.UserIDKey, uint(3))
		c.Set(helpers.UserIsStoreKey, true)
		c.Next()
	}
}

var _ = ginkgo.Describe("offer patch status handler", ginkgo.Ordered, func() {
	dbCont := GetContainer()
	db, err := GetDB(dbCont)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	offerRepo := repository.NewOfferRepository(db)
	offerServ := offer.NewService(offerRepo, nil)
	offerHand := handler.NewOfferHandler(offerServ)

	ginkgo.AfterAll(func() {
		_ = db.Close()
		_ = dbCont.Terminate(context.Background())
	})

	ginkgo.Context("when the user is the shop owner", func() {

		gin.SetMode(gin.ReleaseMode)
		router := gin.New()
		router.Use(middleware.Errors())
		router.Use(mockAuthShopOwnerMiddleware())
		router.PATCH("/api/test/offers/:offerID/status-update", offerHand.PatchOfferStatus)

		ginkgo.It("successfully updates the offer status if everything is fine", func() {
			correctOfferID := 1
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

			var ofr dto.PatchOfferStatusResp
			_ = json.Unmarshal(rec.Body.Bytes(), &ofr)
			gomega.Expect(ofr.NewStatus).To(gomega.Equal("accepted"))
		})

		ginkgo.It("fails data validation if the is negative", func() {
			badOfferID := -2
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", badOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})

		ginkgo.It("fails data validation if the is non-numeric", func() {
			badOfferID := "two"
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%s/status-update", badOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})

		ginkgo.It("fails data validation if the status is not accepted/declined", func() {
			correctOfferID := 4
			badStatus := "bad_status"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: badStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})

		ginkgo.It("fails data validation if the JSON body is malformed", func() {
			correctOfferID := 4
			malformedJSON := []byte(`{"status": "accepted"`)

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", correctOfferID),
				bytes.NewBuffer(malformedJSON))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})

		ginkgo.It("fails if the offer is not found", func() {
			badOfferID := 999
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", badOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNotFound))
		})

		ginkgo.It("fails if the offer is not in a `pending` state", func() {
			correctOfferID := 1 // was changed in the first test to `accepted`
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusConflict))
		})
	})

	ginkgo.Context("when the user is an owner of a different shop", func() {

		router := gin.New()
		gin.SetMode(gin.ReleaseMode)
		router.Use(middleware.Errors())
		router.Use(mockAuthIncorrectShopOwnerMiddleware())
		router.PATCH("/api/test/offers/:offerID/status-update", offerHand.PatchOfferStatus)

		ginkgo.It("fails to update the offer status, even if everything is fine", func() {
			correctOfferID := 2
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))

		})
	})

	ginkgo.Context("when a user is the creator of an offer", func() {

		gin.SetMode(gin.ReleaseMode)
		router := gin.New()

		router.Use(middleware.Errors())
		router.Use(mockAuthBuyerMiddleware())
		router.PATCH("/api/test/offers/:offerID/status-update", offerHand.PatchOfferStatus)

		ginkgo.It("updates the status to `cancelled` if the request is correct", func() {
			correctOfferID := 2
			correctStatus := "cancelled"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

			var ofr dto.PatchOfferStatusResp
			_ = json.Unmarshal(rec.Body.Bytes(), &ofr)
			gomega.Expect(ofr.NewStatus).To(gomega.Equal("cancelled"))
		})

		ginkgo.It("fails to update the status to `accepted`, since that status can only be used by shop owner", func() {
			correctOfferID := 3
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})
	})

	ginkgo.Context("when a user is NOT the creator of an offer", func() {

		gin.SetMode(gin.ReleaseMode)
		router := gin.New()

		router.Use(middleware.Errors())
		router.Use(mockAuthBuyerMiddleware())
		router.PATCH("/api/test/offers/:offerID/status-update", offerHand.PatchOfferStatus)

		ginkgo.It("updates the status to `cancelled` if the request is correct", func() {
			correctOfferID := 5
			correctStatus := "cancelled"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d/status-update", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNotFound))
		})
	})
})
