package mappers

import (
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth/domain"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
)

func ToUserEntity(user domain.User) *entity.User {
	return &entity.User{
		Email:        user.Email,
		PasswordHash: user.HashedPassword,
	}
}
