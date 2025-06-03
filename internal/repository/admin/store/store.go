package store

import (
	"context"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type RepositoryStore interface {
	InsertStore(ctx context.Context, store user.User) error
}

type repositoryStore struct {
	db *sqlx.DB
	l  *zap.Logger
}

func NewRepositoryStore(db *sqlx.DB, l *zap.Logger) RepositoryStore {
	return &repositoryStore{
		db: db,
		l:  l,
	}
}

func (r *repositoryStore) InsertStore(ctx context.Context, store user.User) error {
	storeModel := model.ConvertUserFromSvc(store)

	stmt := sq.Insert("users").
		Columns("name", "email", "phone_number", "password_hash", "is_store").
		Values(store.Name, store.Email, store.Phone, store.Password, store.IsStore).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	query, args := stmt.MustSql()

	err := r.db.QueryRowxContext(ctx, query, args...).Scan(&storeModel.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr); pgErr.Code == pgerrcode.UniqueViolation {
			return apperror.New(apperror.DuplicateError, "user with this email already exists", err)
		}
		return apperror.New(apperror.DuplicateError, "failed to create user", err)
	}

	return nil
}
