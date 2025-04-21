// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token for authentication. Format: "Bearer <token>"

package handler

import (
	"errors"
	"net/http"
	"time"

	_ "github.com/EM-Stawberry/Stawberry/docs"
	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/EM-Stawberry/Stawberry/internal/handler/reviews"
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
	productReviewH *reviews.ProductReviewsHandler,
	sellerReviewH *reviews.SellerReviewsHandler,
	userS middleware.UserGetter,
	tokenS middleware.TokenValidator,
	s3 *objectstorage.BucketBasics,
	basePath string,
	logger *zap.Logger,
) *gin.Engine {
	router := gin.New()

	// Add custom middleware using zap
	router.Use(middleware.ZapLogger(logger))
	router.Use(middleware.ZapRecovery(logger))
	router.Use(middleware.CORS())

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

	// Публичные маршруты (без аутентификации)
	public := base.Group("/")
	{
		public.GET("/products/:id/reviews", productReviewH.GetReviews)
		public.GET("/sellers/:id/reviews", sellerReviewH.GetReviews)
	}

	// Защищённые маршруты (требуется аутентификация)
	secured := base.Group("/")
	secured.Use(middleware.AuthMiddleware(userS, tokenS))
	{
		// Тестовый эндпоинт для проверки аутентификации
		secured.GET("/auth_required", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"time":   time.Now().Unix(),
			})
		})

		// Эндпоинты для добавления отзывов
		secured.POST("/products/:id/reviews", productReviewH.AddReview)
		secured.POST("/sellers/:id/reviews", sellerReviewH.AddReview)
	}

	return router
}

func handleProductError(c *gin.Context, err error) {
	var productErr *apperror.ProductError
	if errors.As(err, &productErr) {
		status := http.StatusInternalServerError

		switch productErr.Code {
		case apperror.NotFound:
			status = http.StatusNotFound
		case apperror.DuplicateError:
			status = http.StatusConflict
		case apperror.DatabaseError:
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{
			"code":    productErr.Code,
			"message": productErr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    apperror.InternalError,
		"message": "An unexpected error occurred",
	})
}

func handleOfferError(c *gin.Context, err error) {
	var offerError *apperror.OfferError
	if errors.As(err, &offerError) {
		status := http.StatusInternalServerError

		switch offerError.Code {
		case apperror.NotFound:
			status = http.StatusNotFound
		case apperror.DuplicateError:
			status = http.StatusConflict
		case apperror.DatabaseError:
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{
			"code":    offerError.Code,
			"message": offerError.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    apperror.InternalError,
		"message": "An unexpected error occurred",
	})
}

func handleUserError(c *gin.Context, err error) {
	var userError *apperror.UserError

	if errors.As(err, &userError) {

		status := http.StatusInternalServerError

		switch userError.Code {
		case apperror.NotFound:
			status = http.StatusNotFound
		case apperror.DuplicateError:
			status = http.StatusConflict
		case apperror.DatabaseError:
			status = http.StatusInternalServerError
		case apperror.Unauthorized:
			status = http.StatusUnauthorized
		}

		c.JSON(status, gin.H{
			"code":    userError.Code,
			"message": userError.Message,
		})
		return
	}

	var tokenError *apperror.TokenError
	if errors.As(err, &tokenError) {
		status := http.StatusInternalServerError

		switch tokenError.Code {
		case apperror.InvalidToken:
			status = http.StatusUnauthorized
		case apperror.NotFound:
			status = http.StatusUnauthorized
		case apperror.InvalidFingerprint:
			status = http.StatusUnauthorized
		}

		c.JSON(status, gin.H{
			"code":    tokenError.Code,
			"message": tokenError.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    apperror.InternalError,
		"message": "An unexpected error occurred",
	})
}

func handleNotificationError(c *gin.Context, err error) {
	var notificationErr *apperror.NotificationError
	if errors.As(err, &notificationErr) {
		status := http.StatusInternalServerError

		// продумать логику ошибок
		switch notificationErr.Code {
		case apperror.NotFound:
			status = http.StatusNotFound
		case apperror.DuplicateError:
			status = http.StatusConflict
		case apperror.DatabaseError:
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{
			"code":    notificationErr.Code,
			"message": notificationErr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    apperror.InternalError,
		"message": "An unexpected error occurred",
	})
}
