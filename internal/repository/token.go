package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type tokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) *tokenRepository {
	return &tokenRepository{db: db}
}

// InsertToken добавляет новый refresh токен в БД.
func (r *tokenRepository) InsertToken(
	ctx context.Context,
	token entity.RefreshToken,
) error {
	stmt := sq.Insert("refresh_tokens").
		Columns("uuid", "created_at", "expires_at", "revoked_at", "fingerprint", "user_id").
		Values(token.UUID, token.CreatedAt, token.ExpiresAt, token.RevokedAt, token.Fingerprint, token.UserID)

	query, args := stmt.PlaceholderFormat(sq.Dollar).MustSql()

	_, err := r.db.ExecContext(ctx, query, args...)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr); pgErr.Code == pgerrcode.UniqueViolation {
			return apperror.New(apperror.DuplicateError, "token with this uuid already exists", err)
		}
		return apperror.New(apperror.DatabaseError, "failed to create token", err)
	}

	return nil
}

// GetActivesTokenByUserID получает список активных refresh токенов пользователя по userID.
func (r *tokenRepository) GetActivesTokenByUserID(
	ctx context.Context,
	userID uint,
) ([]entity.RefreshToken, error) {
	stmt := sq.Select("uuid", "created_at", "expires_at", "revoked_at", "fingerprint", "user_id").
		From("refresh_tokens").
		Where(sq.Eq{"user_id": userID})

	query, args := stmt.PlaceholderFormat(sq.Dollar).MustSql()

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, apperror.New(apperror.DatabaseError, "failed to fetch user tokens", err)
	}

	defer rows.Close()

	tokens := make([]entity.RefreshToken, 0)
	for rows.Next() {
		var tokenModel model.RefreshToken
		if err := rows.StructScan(&tokenModel); err != nil {
			return nil, apperror.New(apperror.DatabaseError, "failed to fetch user tokens", err)
		}
		tokens = append(tokens, model.ConvertTokenToEntity(tokenModel))
	}

	return tokens, nil
}

// RevokeActivesByUserID помечает все активные refresh токены пользователя как отозванные.
func (r *tokenRepository) RevokeActivesByUserID(
	ctx context.Context,
	userID uint,
) error {
	stmt := sq.Update("refresh_tokens").
		Set("revoked_at", sq.Expr("NOW()")).
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"revoked_at": nil})

	query, args := stmt.PlaceholderFormat(sq.Dollar).MustSql()

	res, err := r.db.ExecContext(ctx, query, args...)

	if err != nil {
		return apperror.New(apperror.DatabaseError, "failed to revoke user tokens", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return apperror.New(apperror.DatabaseError, "failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return apperror.ErrTokenNotFound
	}

	return nil
}

// GetByUUID находит refresh токен по его UUID.
func (r *tokenRepository) GetByUUID(
	ctx context.Context,
	uuid string,
) (entity.RefreshToken, error) {
	var tokenModel model.RefreshToken

	stmt := sq.Select("uuid", "created_at", "expires_at", "revoked_at", "fingerprint", "user_id").
		From("refresh_tokens").
		Where(sq.Eq{"uuid": uuid})

	query, args := stmt.PlaceholderFormat(sq.Dollar).MustSql()

	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&tokenModel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.RefreshToken{}, apperror.ErrTokenNotFound
		}
		return entity.RefreshToken{}, apperror.New(apperror.DatabaseError, "failed to fetch token by uuid", err)
	}

	return model.ConvertTokenToEntity(tokenModel), nil
}

// Update обновляет refresh токен.
func (r *tokenRepository) Update(
	ctx context.Context,
	refresh entity.RefreshToken,
) (entity.RefreshToken, error) {
	refreshModel := model.ConvertTokenFromEntity(refresh)

	stmt := sq.Update("refresh_tokens").
		Set("created_at", refresh.CreatedAt).
		Set("expires_at", refresh.ExpiresAt).
		Set("revoked_at", refresh.RevokedAt).
		Set("fingerprint", refresh.Fingerprint).
		Set("user_id", refresh.UserID).
		Where(sq.Eq{"uuid": refresh.UUID})

	query, args := stmt.PlaceholderFormat(sq.Dollar).MustSql()

	res, err := r.db.ExecContext(ctx, query, args...)

	if err != nil {
		return entity.RefreshToken{}, apperror.New(apperror.DatabaseError, "failed to update refresh token", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return entity.RefreshToken{}, apperror.New(apperror.DatabaseError, "failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return entity.RefreshToken{}, apperror.ErrTokenNotFound
	}

	return model.ConvertTokenToEntity(refreshModel), nil
}
