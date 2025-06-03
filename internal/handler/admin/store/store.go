package store

import (
	"net/http"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/admin/store"
	"github.com/EM-Stawberry/Stawberry/internal/handler/admin/dto"
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source=$GOFILE -destination=store_mock_test.go -package=store HandlerStore
type HandlerStore interface {
	Registration(c *gin.Context)
}

type Handler struct {
	refreshLife  int
	adminService store.ServiceStore
}

func NewStoreHandler(cfg *config.Config, adminService store.ServiceStore) HandlerStore {
	return &Handler{
		refreshLife:  int(cfg.Token.RefreshTokenDuration),
		adminService: adminService,
	}
}

// Registration godoc
// @Summary Регистрация нового магазина
// @Description Регистрирует новый магазин
// @Tags admin/store
// @Accept json
// @Produce json
// @Param user body dto.RegisterStoreReq true "Данные для регистрации магазина"
// @Success 201 {object} map[string]string "Store registered successfully"
// @Failure 400 {object} apperror.AppError
// @Router /admin/store [post]
func (h *Handler) Registration(c *gin.Context) {
	var regStoreDTO dto.RegisterStoreReq
	if err := c.ShouldBindJSON(&regStoreDTO); err != nil {
		_ = c.Error(apperror.New(apperror.BadRequest, "Invalid user data", err))
		return
	}

	err := h.adminService.CreateUser(
		c.Request.Context(),
		regStoreDTO.ConvertToSvc(),
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Store registered successfully",
	})
}
