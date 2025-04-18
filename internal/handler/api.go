package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/zuzaaa-dev/stawberry/internal/app/apperror"
	"github.com/zuzaaa-dev/stawberry/internal/handler/middleware"
	productHandler "github.com/zuzaaa-dev/stawberry/internal/handler/product"
	objectstorage "github.com/zuzaaa-dev/stawberry/pkg/s3"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	productH productHandler.ProductHandler,
	offerH offerHandler,
	userH userHandler,
	notificationH notificationHandler,
	s3 *objectstorage.BucketBasics,
	basePath string,
) *gin.Engine {
	router := gin.New()

	// Add default middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	base := router.Group(basePath)
	auth := base.Group("/auth")
	{
		auth.POST("/reg", userH.Registration)
		auth.POST("/login", userH.Login)
		auth.POST("/logout", userH.Logout)
		auth.POST("/refresh", userH.Refresh)
	}

	shop := base.Group("/shop")

	shopAdmin := shop.Group("/admin")
	{
		shopAdmin.POST("/products", productH.PostProduct)
		shopAdmin.GET("/products/:id", productH.GetProduct)
		shopAdmin.GET("/products", productH.GetProducts)
		shopAdmin.GET("/products/:store_id", productH.GetStoreProducts)
		shopAdmin.PATCH("/products/:id", productH.PatchProduct)
		shopAdmin.DELETE("/products/:id", productH.DeleteProduct)
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
