package main

import (
	"log"
	"os"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/notification"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"

	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/EM-Stawberry/Stawberry/pkg/migrator"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/app"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"
	"github.com/EM-Stawberry/Stawberry/internal/handler"
	objectstorage "github.com/EM-Stawberry/Stawberry/pkg/s3"
	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
)

func main() {
	if err := initializeApp(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := app.StartServer(router, port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func initializeApp() error {
	cfg := config.LoadConfig()

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	db := repository.InitDB(cfg)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	migrator.RunMigrations(db, "migrations")

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
	s3 := objectstorage.ObjectStorageConn(cfg)

	router = handler.SetupRouter(productHandler, offerHandler, userHandler, notificationHandler, s3, "api/v1")

	return nil
}
