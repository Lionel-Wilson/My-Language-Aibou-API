package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/volatiletech/null/v8"
	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/subscriptions/storage"
)

type SubscriptionService interface {
	SubscribeUser(ctx context.Context, user *entity.User) (*entity.Subscription, error)
}

type subscriptionService struct {
	logger            *zap.Logger
	stripeSecretKey   string
	subscriptionsRepo storage.SubscriptionsRepository
}

func NewSubscriptionService(
	logger *zap.Logger,
	stripeSecretKey string,
	subscriptionsRepo storage.SubscriptionsRepository,
) SubscriptionService {
	return subscriptionService{
		logger:            logger,
		stripeSecretKey:   stripeSecretKey,
		subscriptionsRepo: subscriptionsRepo,
	}
}

// SubscribeUser creates a new Stripe subscription for the given user and stores it.
func (s subscriptionService) SubscribeUser(ctx context.Context, user *entity.User) (*entity.Subscription, error) {
	// Set the Stripe API key
	stripe.Key = s.stripeSecretKey

	var stripeCustomerID *string
	if user.StripeCustomerID.Valid {
		stripeCustomerID = &user.StripeCustomerID.String
	}

	// Define the subscription parameters (replace "price_xyz" with your Stripe Price ID for £5/month)
	params := &stripe.SubscriptionParams{
		Customer:        stripeCustomerID,
		Items:           []*stripe.SubscriptionItemsParams{{Price: stripe.String("price_1RCJhLGhK964Xz608A8iJO80")}},
		TrialPeriodDays: stripe.Int64(14),
	}

	stripeSub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe subscription: %w", err)
	}

	// Calculate trial start and trial end based on Stripe response.
	// Stripe returns trial end as a Unix timestamp.
	trialEndUnix := stripeSub.TrialEnd
	trialStart := time.Now()
	trialEnd := time.Unix(trialEndUnix, 0)

	// Map the Stripe subscription details to your entity
	subscriptionEntity := &entity.Subscription{
		UserID:               user.ID,
		StripeSubscriptionID: stripeSub.ID,
		Status:               string(stripeSub.Status), // Status can be "trialing", etc.
		TrialStart:           null.TimeFrom(trialStart),
		TrialEnd:             null.TimeFrom(trialEnd),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	createdSub, err := s.subscriptionsRepo.Insert(ctx, subscriptionEntity)
	if err != nil {
		return nil, err
	}

	return createdSub, nil
}
