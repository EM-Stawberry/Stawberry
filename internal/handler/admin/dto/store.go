package dto

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
)

type RegisterStoreReq struct {
	dto.RegistrationUserReq
	IsStore bool `json:"is_store"`
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
