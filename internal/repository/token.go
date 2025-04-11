package repository

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
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

	return nil
}

// GetActivesTokenByUserID получает список активных refresh токенов пользователя по userID.
func (r *tokenRepository) GetActivesTokenByUserID(
	ctx context.Context,
	userID uint,
) ([]entity.RefreshToken, error) {
	var tokensModel []model.RefreshToken

	tokens := make([]entity.RefreshToken, 0, len(tokensModel))
	for _, token := range tokensModel {
		tokens = append(tokens, model.ConvertTokenToEntity(token))
	}

	return tokens, nil
}

// RevokeActivesByUserID помечает все активные refresh токены пользователя как отозванные.
func (r *tokenRepository) RevokeActivesByUserID(
	ctx context.Context,
	userID uint,
) error {

	return nil
}

// GetByUUID находит refresh токен по его UUID.
func (r *tokenRepository) GetByUUID(
	ctx context.Context,
	uuid string,
) (entity.RefreshToken, error) {
	var tokenModel model.RefreshToken

	return model.ConvertTokenToEntity(tokenModel), nil
}

// Update обновляет refresh токен.
func (r *tokenRepository) Update(
	ctx context.Context,
	refresh entity.RefreshToken,
) (entity.RefreshToken, error) {
	refreshModel := model.ConvertTokenFromEntity(refresh)

	return model.ConvertTokenToEntity(refreshModel), nil
}
