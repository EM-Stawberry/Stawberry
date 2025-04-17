package main

import (
	"log"
	"os"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/notification"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/token"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"

	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/EM-Stawberry/Stawberry/pkg/logger"
	"github.com/EM-Stawberry/Stawberry/pkg/migrator"

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
	log := logger.SetupLogger(cfg.Environment)
	log.Info("Config initialized")
	log.Info("Logger initialized")

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Настраиваем Gin на использование Zap логгера
	middleware.SetupGinWithZap(log)

	db := repository.InitDB(cfg)
	log.Info("Database initialized")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	migrator.RunMigrationsWithZap(db, "migrations", log)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	migrator.RunMigrations(db, "migrations")

	productRepository := repository.NewProductRepository(db)
	offerRepository := repository.NewOfferRepository(db)
	userRepository := repository.NewUserRepository(db)
	notificationRepository := repository.NewNotificationRepository(db)
	tokenRepository := repository.NewTokenRepository(db)
	log.Info("Repositories initialized")

	productService := product.NewProductService(productRepository)
	offerService := offer.NewOfferService(offerRepository)
	tokenService := token.NewTokenService(tokenRepository, cfg.Token.Secret, cfg.Token.AccessTokenDuration, cfg.Token.RefreshTokenDuration)
	userService := user.NewUserService(userRepository, tokenService)
	notificationService := notification.NewNotificationService(notificationRepository)
	log.Info("Services initialized")

	productHandler := handler.NewProductHandler(productService)
	offerHandler := handler.NewOfferHandler(offerService)
	userHandler := handler.NewUserHandler(cfg, userService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	log.Info("Handlers initialized")

	router = handler.SetupRouter(
		productHandler,
		offerHandler,
		userHandler,
		notificationHandler,
		userService,
		tokenService,
		"api/v1",
		log,
	)

	return nil
}
