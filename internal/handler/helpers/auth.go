package helpers

import "github.com/gin-gonic/gin"

const (
	UserIDKey      = "userID"
	UserIsStoreKey = "userIsStore"
)

func GetUserID(c *gin.Context) (uint, bool) {
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

func GetUserIsStore(c *gin.Context) (bool, bool) {
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
