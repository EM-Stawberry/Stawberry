package main

import (
	"os"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/notification"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/token"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"go.uber.org/zap"

	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/EM-Stawberry/Stawberry/pkg/database"
	"github.com/EM-Stawberry/Stawberry/pkg/logger"
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

	log := logger.SetupLogger(cfg.Environment)
	log.Info("Config initialized")
	log.Info("Logger initialized")

	if err := initializeApp(cfg, db, log); err != nil {
		log.Fatal("Failed to initialize application",
			zap.Error(err),
		)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := app.StartServer(router, port); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}

func initializeApp(cfg *config.Config, db *sqlx.DB, log *zap.Logger) error {

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Настраиваем Gin на использование Zap логгера
	middleware.SetupGinWithZap(log)

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	migrator.RunMigrationsWithZap(db, "migrations", log)

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
