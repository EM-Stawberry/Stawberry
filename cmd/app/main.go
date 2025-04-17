package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/notification"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/user"

	"github.com/zuzaaa-dev/stawberry/internal/repository"
	"github.com/zuzaaa-dev/stawberry/pkg/database"
	"github.com/zuzaaa-dev/stawberry/pkg/migrator"
	"github.com/zuzaaa-dev/stawberry/pkg/server"

	"github.com/gin-gonic/gin"
	"github.com/zuzaaa-dev/stawberry/config"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/offer"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/product"
	"github.com/zuzaaa-dev/stawberry/internal/handler"
)

var (
	router *gin.Engine
)

func main() {
	cfg := config.LoadConfig()

	db, close := database.InitDB(&cfg.DB)
	defer close()

	migrator.RunMigrations(db, "migrations")

	if err := initializeApp(cfg, db); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := server.StartServer(router, &cfg.Server); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func initializeApp(cfg *config.Config, db *sqlx.DB) error {

	productRepository := repository.NewProductRepository(db)
	offerRepository := repository.NewOfferRepository(db)
	userRepository := repository.NewUserRepository(db)
	notificationRepository := repository.NewNotificationRepository(db)

	productService := product.NewProductService(productRepository)
	offerService := offer.NewOfferService(offerRepository)
	userService := user.NewUserService(userRepository)
	notificationService := notification.NewNotificationService(notificationRepository)

	productHandler := handler.NewProductHandler(productService)
	offerHandler := handler.NewOfferHandler(offerService)
	userHandler := handler.NewUserHandler(cfg, userService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	router = handler.SetupRouter(productHandler, offerHandler, userHandler, notificationHandler, "api/v1")

	return nil
}
