package helpers

import "github.com/gin-gonic/gin"

const (
	UserIDKey      = "userID"
	UserIsStoreKey = "userIsStore"
	UserIsAdminKey = "userIsAdmin"
)

func UserIDContext(c *gin.Context) (uint, bool) {
	id, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	idValue, ok := id.(uint)
	if !ok {
		return 0, false
	}
	return idValue, true
}

func UserIsAdminContext(c *gin.Context) (bool, bool) {
	isAdmin, exists := c.Get(UserIsAdminKey)
	if !exists {
		return false, false
	}
	isAdminValue, ok := isAdmin.(bool)
	if !ok {
		return false, false
	}
	return isAdminValue, true
}

func UserIsStoreContext(c *gin.Context) (bool, bool) {
	isStore, exists := c.Get(UserIsStoreKey)
	if !exists {
		return false, false
	}
	isStoreValue, ok := isStore.(bool)
	if !ok {
		return false, false
	}
	return isStoreValue, true
}
