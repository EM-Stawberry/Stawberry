package middleware

import (
	"net/http"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"
	"github.com/gin-gonic/gin"
)

func CheckAdminRights() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, ok := helpers.UserIsAdminContext(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, apperror.ErrUserNotFound)
			c.Abort()
		}
		if !isAdmin {
			c.JSON(http.StatusForbidden, apperror.ErrNoPermissions)
			c.Abort()
		}

		c.Next()
	}
}
