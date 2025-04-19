package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/user"
	"github.com/zuzaaa-dev/stawberry/internal/repository/model"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

// InsertUser вставляет пользователя в БД
func (r *userRepository) InsertUser(
	ctx context.Context,
	user user.User,
) (uint, error) {

	userModel := model.ConvertUserFromSvc(user)
	return userModel.ID, nil
}

// GetUser получает пользователя по почте
func (r *userRepository) GetUser(
	ctx context.Context,
	email string,
) (entity.User, error) {
	var userModel model.User

	return model.ConvertUserToEntity(userModel), nil
}

// GetUserByID получает пользователя по айди
func (r *userRepository) GetUserByID(
	ctx context.Context,
	id uint,
) (entity.User, error) {
	var userModel model.User

	return model.ConvertUserToEntity(userModel), nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user user.User) error {
	panic("implement me")
}
