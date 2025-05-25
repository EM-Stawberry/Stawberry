// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token for authentication. Format: "Bearer <token>"

package handler

import (

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"golang.org/x/text/currency"

	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"
	// Импорт сваггер-генератора
	_ "github.com/EM-Stawberry/Stawberry/docs"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/EM-Stawberry/Stawberry/internal/handler/reviews"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary Получить статус сервера
// @Description Возвращает статус сервера и текущее время
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Успешный ответ с данными"
// @Router /health [get]
func SetupRouter(
	healthH *HealthHandler,
	productH *ProductHandler,
	offerH *OfferHandler,
	userH *UserHandler,
	notificationH *NotificationHandler,
	productReviewH *reviews.ProductReviewsHandler,
	sellerReviewH *reviews.SellerReviewsHandler,
	userS middleware.UserGetter,
	tokenS middleware.TokenValidator,
	basePath string,
	logger *zap.Logger,
) *gin.Engine {
	router := gin.New()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("iso4217", currencyValidator)
	}


	// Add custom middleware using zap
	router.Use(middleware.ZapLogger(logger))
	router.Use(middleware.ZapRecovery(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Errors())

	// Swagger UI эндпоинт
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// base это эндпойнты без префикса версии
	base := router.Group(basePath)

	// public это эндпойнты с префиксом версии
	public := base.Group("/")

	// secured это эндпойнты, которые не сработают без авторизационного токера
	secured := base.Use(middleware.AuthMiddleware(userS, tokenS))

	// healtcheck эндпойнты
	{
		base.GET("/health", healthH.health)
		secured.GET("/auth_required", healthH.authCheck)
	}

	// эндпойнты регистрации-авторизации
	auth := public.Group("/auth")
	{
		auth.POST("/reg", userH.Registration)
		auth.POST("/login", userH.Login)
		auth.POST("/logout", userH.Logout)
		auth.POST("/refresh", userH.Refresh)
	}

	public := base.Group("/")
	{

		public.GET("/products/:id/reviews", productReviewH.GetReviews)
		public.GET("/sellers/:id/reviews", sellerReviewH.GetReviews)
	}

	secured := base.Use(middleware.AuthMiddleware(userS, tokenS))
	{
		// Тестовый эндпоинт для проверки аутентификации
		secured.GET("/auth_required", func(c *gin.Context) {
			userID, ok := helpers.UserIDContext(c)
			var status string
			if ok {
				status = "UserID found"
			} else {
				status = "UserID not found"
			}
			isStore, ok := helpers.UserIsStoreContext(c)

			if !ok {
				logger.Warn("Missing isStore field in context")
			}

			c.JSON(http.StatusOK, gin.H{
				"userID":  userID,
				"status":  status,
				"isStore": isStore,
				"time":    time.Now().Unix(),
			})
		})

		secured.PATCH("offers/:offerID", offerH.PatchOfferStatus)
		secured.POST("offers", offerH.PostOffer)

		// Эндпоинты для добавления отзывов
		secured.POST("/products/:id/reviews", productReviewH.AddReview)
		secured.POST("/sellers/:id/reviews", sellerReviewH.AddReview)
	}

	// Эти заглушки можно убрать после реализации соответствующих хендлеров
	_ = productH
	_ = notificationH

	return router
}

var currencyValidator validator.Func = func(fl validator.FieldLevel) bool {
	currencyCode := fl.Field().String()
	_, err := currency.ParseISO(currencyCode)
	return err == nil
}
