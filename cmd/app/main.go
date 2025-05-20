package main

import (
	flag "github.com/spf13/pflag"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/notification"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/token"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"go.uber.org/zap"

	"github.com/EM-Stawberry/Stawberry/internal/adapter/auth"
	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/EM-Stawberry/Stawberry/pkg/database"
	"github.com/EM-Stawberry/Stawberry/pkg/email"
	"github.com/EM-Stawberry/Stawberry/pkg/logger"
	"github.com/EM-Stawberry/Stawberry/pkg/migrator"
	"github.com/EM-Stawberry/Stawberry/pkg/security"
	"github.com/EM-Stawberry/Stawberry/pkg/server"
	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"
	"github.com/EM-Stawberry/Stawberry/internal/handler"
	"github.com/gin-gonic/gin"
)

var enableMail bool

func init() {
	flag.BoolVarP(&enableMail, "mail", "m", false, "enable email notifications")
}

func main() {
	flag.Parse()

	cfg := config.LoadConfig()
	log := logger.SetupLogger(cfg.Environment)
	log.Info("Logger initialized")

	db, closer := database.InitDB(&cfg.DB)
	defer closer()

	migrator.RunMigrationsWithZap(db, "migrations", log)

	router, mailer := initializeApp(cfg, db, log)

	if err := server.StartServer(router, mailer, &cfg.Server); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}

func initializeApp(cfg *config.Config, db *sqlx.DB, log *zap.Logger) (*gin.Engine, email.MailerService) {
	mailer := email.NewMailer(log, &cfg.Email)
	log.Info("Mailer initialized")

	productRepository := repository.NewProductRepository(db)
	offerRepository := repository.NewOfferRepository(db)
	userRepository := repository.NewUserRepository(db)
	notificationRepository := repository.NewNotificationRepository(db)
	tokenRepository := repository.NewTokenRepository(db)
	log.Info("Repositories initialized")

	passwordManager := security.NewArgon2idPasswordManager()
	jwtManager := auth.NewJWTManager(cfg.Token.Secret)

	productService := product.NewService(productRepository)
	offerService := offer.NewService(offerRepository, mailer)
	tokenService := token.NewService(
		tokenRepository,
		jwtManager,
		cfg.Token.RefreshTokenDuration,
		cfg.Token.AccessTokenDuration,
	)
	userService := user.NewService(userRepository, tokenService, passwordManager, mailer)
	notificationService := notification.NewService(notificationRepository)
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

	return router, mailer
}
