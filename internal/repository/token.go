package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
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

	query := `INSERT into refresh_tokens (uuid, created_at, expires_at, revoked_at, fingerprint, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query, token.UUID, token.CreatedAt, token.ExpiresAt, token.RevokedAt, token.Fingerprint, token.UserID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr); pgErr.Code == pgerrcode.UniqueViolation {
			return &apperror.TokenError{
				Code:    apperror.DuplicateError,
				Message: "token with this uuid already exists",
				Err:     err,
			}
		}
		return &apperror.TokenError{
			Code:    apperror.DatabaseError,
			Message: "failed to create token",
			Err:     err,
		}
	}

	return nil
}

// GetActivesTokenByUserID получает список активных refresh токенов пользователя по userID.
func (r *tokenRepository) GetActivesTokenByUserID(
	ctx context.Context,
	userID uint,
) ([]entity.RefreshToken, error) {
	query := `SELECT uuid, created_at, expires_at, revoked_at, fingerprint, user_id
		FROM refresh_tokens WHERE user_id = $1`

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, &apperror.TokenError{
			Code:    apperror.DatabaseError,
			Message: "failed to fetch user tokens",
			Err:     err,
		}
	}

	defer rows.Close()

	tokens := make([]entity.RefreshToken, 0)
	for rows.Next() {
		var tokenModel model.RefreshToken
		if err := rows.StructScan(&tokenModel); err != nil {
			return nil, &apperror.TokenError{
				Code:    apperror.DatabaseError,
				Message: "failed to fetch user tokens",
				Err:     err,
			}
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

	query := `UPDATE refresh_tokens
						SET revoked_at=NOW()
						WHERE user_id=$1 and revoked_at is NULL`
	res, err := r.db.ExecContext(ctx, query, userID)

	if err != nil {
		return &apperror.TokenError{
			Code:    apperror.DatabaseError,
			Message: "failed to revoke user tokens",
			Err:     err,
		}
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return &apperror.TokenError{
			Code:    apperror.DatabaseError,
			Message: "failed to get rows affected",
			Err:     err,
		}
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

	query := `SELECT uuid, created_at, expires_at, revoked_at, fingerprint, user_id
		FROM refresh_tokens WHERE uuid = $1`
	err := r.db.QueryRowxContext(ctx, query, uuid).StructScan(&tokenModel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.RefreshToken{}, apperror.ErrTokenNotFound
		}
		return entity.RefreshToken{}, &apperror.TokenError{
			Code:    apperror.DatabaseError,
			Message: "failed to fetch token by uuid",
			Err:     err,
		}
	}

	return model.ConvertTokenToEntity(tokenModel), nil
}

// Update обновляет refresh токен.
func (r *tokenRepository) Update(
	ctx context.Context,
	refresh entity.RefreshToken,
) (entity.RefreshToken, error) {
	refreshModel := model.ConvertTokenFromEntity(refresh)

	query := `UPDATE refresh_tokens
		SET created_at = $1, expires_at = $2, revoked_at = $3, fingerprint = $4, user_id = $5
		WHERE uuid = $6
		RETURNING uuid, created_at, expires_at, revoked_at, fingerprint, user_id`

	res, err := r.db.ExecContext(ctx, query,
		refresh.CreatedAt, refresh.ExpiresAt, refresh.RevokedAt,
		refresh.Fingerprint, refresh.UserID, refresh.UUID)

	if err != nil {
		return entity.RefreshToken{}, &apperror.TokenError{
			Code:    apperror.DatabaseError,
			Message: "failed to update refresh token",
			Err:     err,
		}
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return entity.RefreshToken{}, &apperror.TokenError{
			Code:    apperror.DatabaseError,
			Message: "failed to get rows affected",
			Err:     err,
		}
	}

	if rowsAffected == 0 {
		return entity.RefreshToken{}, apperror.ErrTokenNotFound
	}

	return model.ConvertTokenToEntity(refreshModel), nil
}
