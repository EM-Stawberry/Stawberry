package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/EM-Stawberry/Stawberry/internal/handler"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

type CleanupFunc func(ctx context.Context, opts ...testcontainers.TerminateOption) error

var (
	imageName  = "postgres:17.4-alpine"
	dbName     = "postgres"
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

	return pgContainer
}

func GetDB() (*sqlx.DB, *postgres.PostgresContainer, CleanupFunc, error) {
	pgContainer := GetContainer()
	resp, err := pgContainer.Inspect(context.Background())
	fmt.Println(resp.State.Status)

	connString, err := pgContainer.ConnectionString(context.Background(), "sslmode=disable")
	if err != nil {
		slog.Error(err.Error())
		pgContainer.Terminate(context.Background())
		return nil, nil, nil, err
	}

	db, err := sqlx.Connect("pgx", connString)
	if err != nil {
		pgContainer.Terminate(context.Background())
		return nil, nil, nil, err
	}

	//migrator.RunMigrations(db, "migrations")
	goose.SetDialect("postgres")

	os.Chdir("..")
	os.Chdir("..")

	err = goose.Up(db.DB, `migrations`)
	if err != nil {
		pgContainer.Terminate(context.Background())
		slog.Error(err.Error())
		return nil, nil, nil, err
	}

	_, err = sqlx.LoadFile(db, `.\internal\handler\testdata\offer\sql\populate_test_db.sql`)
	if err != nil {
		slog.Error(err.Error())
		return nil, nil, nil, err
	}

	return db, pgContainer, pgContainer.Terminate, nil
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
		c.Next()
	}
}

var _ = Describe("offer patch status handler", Ordered, func() {
	db, container, cleanup, _ := GetDB()
	_ = cleanup

	container.Snapshot(context.Background())

	offerRepo := repository.NewOfferRepository(db)
	offerServ := offer.NewOfferService(offerRepo)
	offerHand := handler.NewOfferHandler(offerServ)

	Context("when the user is the shop owner", func() {

		router := gin.Default()
		router.Use(middleware.Errors())
		router.Use(mockAuthShopOwnerMiddleware())
		router.PATCH("/api/test/offers/:offerID/status-update", offerHand.PatchOfferStatus)

		It("successfully updates the offer status if everything is fine", func() {
			container.Restore(context.Background())

			correctOfferID := 4
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

			Expect(rec.Code).To(Equal(http.StatusOK))

			var ofr entity.Offer
			_ = json.Unmarshal(rec.Body.Bytes(), &ofr)
			Expect(ofr.Status).To(Equal("accepted"))
		})

		It("fails data validation if the is negative", func() {
			container.Restore(context.Background())

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

			Expect(rec.Code).To(Equal(http.StatusBadRequest))
		})

		It("fails data validation if the is non-numeric", func() {
			container.Restore(context.Background())

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

			Expect(rec.Code).To(Equal(http.StatusBadRequest))
		})

		It("fails data validation if the status is not accepted/declined", func() {
			container.Restore(context.Background())

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

			Expect(rec.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("when the user is an owner of a different shop", func() {

		router := gin.Default()
		router.Use(middleware.Errors())
		router.Use(mockAuthIncorrectShopOwnerMiddleware())
		router.PATCH("/api/test/offers/:offerID/status-update", offerHand.PatchOfferStatus)

		It("fails to update the offer status, even if everything is fine", func() {
			container.Restore(context.Background())

			correctOfferID := 4
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

			Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})
