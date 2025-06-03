package admin

import (
	"github.com/EM-Stawberry/Stawberry/internal/repository/admin/store"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type RepositoryAdmin struct {
	store.RepositoryStore
}

func NewRepository(db *sqlx.DB, l *zap.Logger) *RepositoryAdmin {
	return &RepositoryAdmin{
		store.NewRepositoryStore(db, l),
	}
}
