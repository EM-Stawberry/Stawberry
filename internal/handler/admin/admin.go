package admin

import (
	"github.com/EM-Stawberry/Stawberry/config"
	adminService "github.com/EM-Stawberry/Stawberry/internal/domain/service/admin"
	"github.com/EM-Stawberry/Stawberry/internal/handler/admin/store"
)

type Handler struct {
	store.HandlerStore
}

func NewAdminHandler(cfg *config.Config, adminS *adminService.Service) *Handler {
	return &Handler{
		store.NewStoreHandler(cfg, adminS),
	}
}
