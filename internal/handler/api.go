// @title Stawberry API
// @version 1.0
// @description Это API для управления сделаками по продуктам.
// @host localhost:8080
// @BasePath /

package handler

import (
	"net/http"
	"time"

	_ "github.com/EM-Stawberry/Stawberry/docs"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	objectstorage "github.com/EM-Stawberry/Stawberry/pkg/s3"
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
	productH productHandler,
	offerH offerHandler,
	userH userHandler,
	notificationH notificationHandler,
	userS middleware.UserGetter,
	tokenS middleware.TokenValidator,
	s3 *objectstorage.BucketBasics,
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

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// Swagger UI endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	base := router.Group(basePath)
	auth := base.Group("/auth")
	{
		auth.POST("/reg", userH.Registration)
		auth.POST("/login", userH.Login)
		auth.POST("/logout", userH.Logout)
		auth.POST("/refresh", userH.Refresh)
	}

	secured := base.Use(middleware.AuthMiddleware(userS, tokenS))
	{
		secured.GET("/auth_required", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"time":   time.Now().Unix(),
			})
		})
	}

	return router
}
