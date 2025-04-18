package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	userModel := model.ConvertUserFromSvc(user)

	query := `INSERT INTO users (name, email, phone_number, password_hash, is_store)
					VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := r.db.QueryRowxContext(ctx, query,
		user.Name,
		user.Email,
		user.Phone,
		user.Password,
		user.IsStore,
	).Scan(&userModel.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr); pgErr.Code == pgerrcode.UniqueViolation {
			return 0, &apperror.UserError{
				Code:    apperror.DuplicateError,
				Message: "user with this email already exists",
				Err:     err,
			}
		}
		return 0, &apperror.UserError{
			Code:    apperror.DatabaseError,
			Message: "failed to create user",
			Err:     err,
		}
	}

	return userModel.ID, nil
}

// GetUser получает пользователя по почте
func (r *userRepository) GetUser(
	ctx context.Context,
	email string,
) (entity.User, error) {
	var userModel model.User

	query := `SELECT id, name, email, phone_number, password_hash, is_store
				FROM users WHERE email=$1`
	err := r.db.QueryRowxContext(ctx, query, email).StructScan(&userModel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, apperror.ErrUserNotFound
		}
		return entity.User{}, &apperror.UserError{
			Code:    apperror.DatabaseError,
			Message: "failed to fetch user by email",
			Err:     err,
		}
	}

	return model.ConvertUserToEntity(userModel), nil
}

// GetUserByID получает пользователя по айди
func (r *userRepository) GetUserByID(
	ctx context.Context,
	id uint,
) (entity.User, error) {
	var userModel model.User

	query := `SELECT id, name, email, phone_number, password_hash, is_store
				FROM users WHERE id=$1`
	err := r.db.QueryRowxContext(ctx, query, id).StructScan(&userModel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, apperror.ErrUserNotFound
		}
		return entity.User{}, &apperror.UserError{
			Code:    apperror.DatabaseError,
			Message: "failed to fetch user by ID",
			Err:     err,
		}
	}

	return model.ConvertUserToEntity(userModel), nil
}
