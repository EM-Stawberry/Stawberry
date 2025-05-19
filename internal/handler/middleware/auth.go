package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"

	"github.com/gin-gonic/gin"
)

type UserGetter interface {
	GetUserByID(ctx context.Context, id uint) (entity.User, error)
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (entity.AccessToken, error)
}

// AuthMiddleware валидирует access token,
// достает из него userID и проверяет существование пользователя
func AuthMiddleware(userGetter UserGetter, validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHead := c.GetHeader("Authorization")
		if authHead == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    apperror.Unauthorized,
				"message": "Authorization header is missing",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHead, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    apperror.Unauthorized,
				"message": "Invalid authorization format",
			})
			c.Abort()
			return
		}

		access, err := validator.ValidateToken(c, parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    apperror.Unauthorized,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		user, err := userGetter.GetUserByID(context.Background(), access.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    apperror.Unauthorized,
				"message": "User not found",
			})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
