package admin

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/admin/store"
	"github.com/EM-Stawberry/Stawberry/internal/repository/admin"
)

type Service struct {
	store.ServiceStore
}

func NewAdminService(repo *admin.RepositoryAdmin) *Service {
	return &Service{
		store.NewStoreService(repo),
	}
}
