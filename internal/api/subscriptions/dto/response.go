package dto

import (
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
	"time"
)

type StatusResponse struct {
	Id                   string      `json:"id"`
	UserId               string      `json:"userID"`
	StripeSubscriptionId string      `json:"stripeSubscriptionID"`
	Status               string      `json:"status"`
	TrialStart           time.Time   `json:"trialStart"`
	TrialEnd             time.Time   `json:"trialEnd"`
	StartedAt            interface{} `json:"startedAt"`
	NextBillingDate      interface{} `json:"nextBillingDate"`
	CreatedAt            time.Time   `json:"createdAt"`
	UpdatedAt            time.Time   `json:"updatedAt"`
}

func ToStatusResponse(subscription *entity.Subscription) (*StatusResponse, error) {
	var trialStart time.Time
	if subscription.TrialStart.Valid {
		trialStart = subscription.TrialStart.Time
	}

	var trialEnd time.Time
	if subscription.TrialEnd.Valid {
		trialEnd = subscription.TrialEnd.Time
	}

	var startedAt interface{}
	if subscription.StartedAt.Valid {
		startedAt = subscription.StartedAt.Time
	}

	var nextBillingDate interface{}
	if subscription.NextBillingDate.Valid {
		nextBillingDate = subscription.NextBillingDate.Time
	}

	return &StatusResponse{
		Id:                   subscription.ID,
		UserId:               subscription.UserID,
		StripeSubscriptionId: subscription.StripeSubscriptionID,
		Status:               subscription.Status,
		TrialStart:           trialStart,
		TrialEnd:             trialEnd,
		StartedAt:            startedAt,
		NextBillingDate:      nextBillingDate,
		CreatedAt:            subscription.CreatedAt,
		UpdatedAt:            subscription.UpdatedAt,
	}, nil
}
