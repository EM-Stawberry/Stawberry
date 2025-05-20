package main

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/notification"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/token"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"go.uber.org/zap"

	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/EM-Stawberry/Stawberry/pkg/database"
	"github.com/EM-Stawberry/Stawberry/pkg/logger"
	"github.com/EM-Stawberry/Stawberry/pkg/migrator"
	"github.com/EM-Stawberry/Stawberry/pkg/server"
	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"
	"github.com/EM-Stawberry/Stawberry/internal/handler"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.LoadConfig()
	log := logger.SetupLogger(cfg.Environment)
	log.Info("Logger initialized")

	db, close := database.InitDB(&cfg.DB)
	defer close()

	migrator.RunMigrationsWithZap(db, "migrations", log)

	router := initializeApp(cfg, db, log)

	if err := server.StartServer(router, &cfg.Server); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}

func initializeApp(cfg *config.Config, db *sqlx.DB, log *zap.Logger) *gin.Engine {

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

	healthHandler := handler.NewHealthHandler()
	productHandler := handler.NewProductHandler(productService)
	offerHandler := handler.NewOfferHandler(offerService)
	userHandler := handler.NewUserHandler(cfg, userService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	log.Info("Handlers initialized")

	router := handler.SetupRouter(
		healthHandler,
		productHandler,
		offerHandler,
		userHandler,
		notificationHandler,
		userService,
		tokenService,
		"api/v1",
		log,
	)

	return router
}
