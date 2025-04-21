package repository

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type notificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *notificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) SelectUserNotifications(
	id string,
	offset, limit int,
) ([]entity.Notification, int, error) {
	var total int64

	return nil, int(total), nil
}
