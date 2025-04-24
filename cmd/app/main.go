package main

import (
	"log"
	"os"
	"time"

	"github.com/zuzaaa-dev/stawberry/internal/domain/service/notification"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/user"
	"github.com/zuzaaa-dev/stawberry/internal/handler/middleware"

	"github.com/zuzaaa-dev/stawberry/internal/repository"
	"github.com/zuzaaa-dev/stawberry/pkg/logger"
	"github.com/zuzaaa-dev/stawberry/pkg/migrator"

	"github.com/gin-gonic/gin"
	"github.com/zuzaaa-dev/stawberry/config"
	"github.com/zuzaaa-dev/stawberry/internal/app"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/offer"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/product"
	"github.com/zuzaaa-dev/stawberry/internal/handler"
	objectstorage "github.com/zuzaaa-dev/stawberry/pkg/s3"
)

// Global variables for application state
var (
	router *gin.Engine
)

func main() {
	// Initialize application
	if err := initializeApp(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	if err := app.StartServer(router, port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// initializeApp initializes all application components
func initializeApp() error {
	// Load configuration
	cfg := config.LoadConfig()
	log := logger.SetupLogger(cfg.Environment)
	log.Info("Config initialized")
	log.Info("Logger initialized")

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Настраиваем Gin на использование Zap логгера
	middleware.SetupGinWithZap(log)

	// Initialize database connection
	db := repository.InitDB(cfg)
	log.Info("Connection initialized")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	// Run migrations using Zap logger
	migrator.RunMigrationsWithZap(db, "migrations", log)

	productRepository := repository.NewProductRepository(db)
	offerRepository := repository.NewOfferRepository(db)
	userRepository := repository.NewUserRepository(db)
	notificationRepository := repository.NewNotificationRepository(db)
	log.Info("Repositories initialized")

	productService := product.NewProductService(productRepository)
	offerService := offer.NewOfferService(offerRepository)
	userService := user.NewUserService(userRepository)
	notificationService := notification.NewNotificationService(notificationRepository)
	log.Info("Services initialized")

	productHandler := handler.NewProductHandler(productService)
	offerHandler := handler.NewOfferHandler(offerService)
	userHandler := handler.NewUserHandler(userService, time.Hour, "api/v1", "")
	notificationHandler := handler.NewNotificationHandler(notificationService)
	log.Info("Handlers initialized")

	s3 := objectstorage.ObjectStorageConn(cfg)
	log.Info("Storage initialized")

	router = handler.SetupRouter(
		productHandler,
		offerHandler,
		userHandler,
		notificationHandler,
		s3,
		"api",
		log,
	)

	return nil
}
