package dto

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
)

type RegisterStoreReq struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	IsStore  bool   `json:"is_store"`
}

func (ru *RegisterStoreReq) ConvertToSvc() user.User {
	return user.User{
		Name:     ru.Name,
		Password: ru.Password,
		Email:    ru.Email,
		Phone:    ru.Phone,
		IsStore:  ru.IsStore,
	}
}
