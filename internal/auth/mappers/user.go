package mappers

import (
	"github.com/stripe/stripe-go/v82"
	"github.com/volatiletech/null/v8"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth/domain"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
)

func ToUserEntity(user domain.User, stripeCustomer *stripe.Customer) *entity.User {
	return &entity.User{
		Email:            user.Email,
		PasswordHash:     user.HashedPassword,
		StripeCustomerID: null.StringFrom(stripeCustomer.ID),
	}
}
