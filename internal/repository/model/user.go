package model

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
)

type User struct {
	ID            uint   `gorm:"column:id"`
	Name          string `gorm:"column:name"`
	Email         string `gorm:"column:email"`
	Phone         string `gorm:"column:phone_number"`
	Password      string `gorm:"column:password_hash"`
	IsStore       bool   `gorm:"column:is_store"`
	Notifications []Notification
}

func ConvertUserFromSvc(u user.User) User {
	return User{
		Name:     u.Name,
		Email:    u.Email,
		Phone:    u.Phone,
		Password: u.Password,
		IsStore:  u.IsStore,
	}
}

func ConvertUserToEntity(u User) entity.User {
	return entity.User{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Phone:    u.Phone,
		Password: u.Password,
		IsStore:  u.IsStore,
	}
}
