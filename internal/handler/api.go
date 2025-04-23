package handler

import (
	"net/http"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	objectstorage "github.com/EM-Stawberry/Stawberry/pkg/s3"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	productH productHandler,
	offerH offerHandler,
	userH userHandler,
	notificationH notificationHandler,
	userS middleware.UserGetter,
	tokenS middleware.TokenValidator,
	s3 *objectstorage.BucketBasics,
	basePath string,
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Errors())

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
