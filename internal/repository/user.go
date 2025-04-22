package repository

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
	"github.com/jmoiron/sqlx"
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
	// userModel := model.ConvertUserFromSvc(user)
	// if err := r.db.WithContext(ctx).Create(&userModel).Error; err != nil {
	// 	if isDuplicateError(err) {
	// 		return 0, &apperror.UserError{
	// 			Code:    apperror.DuplicateError,
	// 			Message: "user with this email already exists",
	// 			Err:     err,
	// 		}
	// 	}
	// 	return 0, &apperror.UserError{
	// 		Code:    apperror.DatabaseError,
	// 		Message: "failed to create user",
	// 		Err:     err,
	// 	}
	// }
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
