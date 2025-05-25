// @title Stawberry API
// @version 1.0
// @description Это API для управления сделаками по продуктам.
// @host localhost:8080
// @BasePath /

package handler

import (
	"net/http"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"
	// Импорт сваггер-генератора
	_ "github.com/EM-Stawberry/Stawberry/docs"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
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
	userS middleware.UserGetter,
	tokenS middleware.TokenValidator,
	basePath string,
	logger *zap.Logger,
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	// Add custom middleware using zap
	router.Use(middleware.ZapLogger(logger))
	router.Use(middleware.ZapRecovery(logger))

	router.Use(middleware.CORS())
	router.Use(middleware.Errors())

	// Swagger UI endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	healthH.RegisterRoutes(router)

	base := router.Group(basePath)

	auth := base.Group("/auth")
	userH.RegisterRoutes(auth)

	// Заглушки для нереализованных хендлеров.
	// Не забудьте убрать их и добавить вызов .RegisterRoutes для каждого хендлера
	_ = productH
	_ = offerH
	_ = notificationH

	secured := base.Use(middleware.AuthMiddleware(userS, tokenS))
	{
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
	}

	return router
}
