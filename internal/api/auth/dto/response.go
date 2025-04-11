package dto

import (
	"time"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
)

type UserDetailsResponse struct {
	Email      string    `json:"email,omitempty" validate:"omitempty,email"`
	Status     string    `json:"status,omitempty" `
	TrialStart time.Time `json:"trialStart,omitempty"`
	TrialEnd   time.Time `json:"trialEnd,omitempty"`
}

func ToUserDetailsResponse(user *entity.User, subscription *entity.Subscription) UserDetailsResponse {
	var trialStart time.Time
	if subscription.TrialStart.Valid {
		trialStart = subscription.TrialStart.Time
	}

	var trialEnd time.Time
	if subscription.TrialEnd.Valid {
		trialEnd = subscription.TrialEnd.Time
	}

	return UserDetailsResponse{
		Email:      user.Email,
		Status:     subscription.Status,
		TrialStart: trialStart,
		TrialEnd:   trialEnd,
	}
}

type RegisterResponse struct {
	Email      string    `json:"email,omitempty"`
	Status     string    `json:"status,omitempty" `
	TrialStart time.Time `json:"trialStart,omitempty"`
	TrialEnd   time.Time `json:"trialEnd,omitempty"`
}

func ToRegisterResponse(user *entity.User, subscription *entity.Subscription) *RegisterResponse {
	var trialStart time.Time
	if subscription.TrialStart.Valid {
		trialStart = subscription.TrialStart.Time
	}

	var trialEnd time.Time
	if subscription.TrialEnd.Valid {
		trialEnd = subscription.TrialEnd.Time
	}

	return &RegisterResponse{
		Email:      user.Email,
		Status:     subscription.Status,
		TrialStart: trialStart,
		TrialEnd:   trialEnd,
	}
}
