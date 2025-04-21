// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token for authentication. Format: "Bearer <token>"

package handler

import (
	"net/http"
	"time"

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

	// router.Use(gin.Logger())
	// router.Use(gin.Recovery())

	// Add custom middleware using zap
	router.Use(middleware.ZapLogger(logger))
	router.Use(middleware.ZapRecovery(logger))

	router.Use(middleware.CORS())
	router.Use(middleware.Errors())

	// Эндпоинт для проверки здоровья сервиса
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// Swagger UI эндпоинт
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	base := router.Group(basePath)
	auth := base.Group("/auth")
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

		// Эндпоинты для добавления отзывов
		secured.POST("/products/:id/reviews", productReviewH.AddReview)
		secured.POST("/sellers/:id/reviews", sellerReviewH.AddReview)
	}

	return router
}
