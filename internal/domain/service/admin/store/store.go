package store

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/repository/admin/store"
)

type ServiceStore interface {
	CreateUser(ctx context.Context, user user.User) error
}

type Store struct {
	repo store.RepositoryStore
}

func NewStoreService(repo store.RepositoryStore) *Store {
	return &Store{
		repo: repo,
	}
}

func (s *Store) CreateUser(ctx context.Context, user user.User) error {
	return s.repo.InsertStore(ctx, user)
}
