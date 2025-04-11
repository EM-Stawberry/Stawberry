// @title Stawberry API
// @version 1.0
// @description Это API для управления сделаками по продуктам.
// @host localhost:8080
// @BasePath /

package handler

import (
	"errors"
	"net/http"
	"time"

	_ "github.com/EM-Stawberry/Stawberry/docs"
	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
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
	s3 *objectstorage.BucketBasics,
	basePath string,
	logger *zap.Logger,
) *gin.Engine {
	router := gin.New()

	// Add custom middleware using zap
	router.Use(middleware.ZapLogger(logger))
	router.Use(middleware.ZapRecovery(logger))
	router.Use(middleware.CORS())

	// Health check endpoint
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
		}

		c.JSON(status, gin.H{
			"code":    userError.Code,
			"message": userError.Message,
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
