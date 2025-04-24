package middleware

import (
	"context"
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
			c.Error(apperror.New(apperror.Unauthorized, "Authorization header is missing", nil))
			c.Abort()
			return
		}

		parts := strings.Split(authHead, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Error(apperror.New(apperror.Unauthorized, "Invalid authorization format", nil))
			c.Abort()
			return
		}

		access, err := validator.ValidateToken(c, parts[1])
		if err != nil {
			c.Error(apperror.New(apperror.Unauthorized, "Invalid or expired token", err))
			c.Abort()
			return
		}

		user, err := userGetter.GetUserByID(c.Request.Context(), access.UserID)
		if err != nil {
			c.Error(apperror.New(apperror.Unauthorized, "User not found", err))
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
