package main

import (
	"log"
	"os"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/notification"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"

	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/EM-Stawberry/Stawberry/pkg/database"
	"github.com/EM-Stawberry/Stawberry/pkg/migrator"
	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/app"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"
	"github.com/EM-Stawberry/Stawberry/internal/handler"
	"github.com/gin-gonic/gin"
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

	if err := app.StartServer(router, cfg.Server.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func initializeApp(cfg *config.Config, db *sqlx.DB) error {

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

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
