package notification

import "github.com/EM-Stawberry/Stawberry/internal/domain/entity"

type Repository interface {
	SelectUserNotifications(id string, offset, limit int) ([]entity.Notification, int, error)
}

type NotificationService struct {
	notificationRepository Repository
}

func NewNotificationService(notificationRepository Repository) *NotificationService {
	return &NotificationService{notificationRepository}
}

func (ns *NotificationService) GetNotification(id string, offset int, limit int) ([]entity.Notification, int, error) {
	return ns.notificationRepository.SelectUserNotifications(id, offset, limit)
}
