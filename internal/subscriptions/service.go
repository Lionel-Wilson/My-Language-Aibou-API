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
	GetUserSubscription(ctx context.Context, userID *string) (*entity.Subscription, error)
	CancelSubscription(ctx context.Context, useruserID *string) (*entity.Subscription, error)
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
	return &subscriptionService{
		logger:            logger,
		stripeSecretKey:   stripeSecretKey,
		subscriptionsRepo: subscriptionsRepo,
	}
}

func (s *subscriptionService) GetUserSubscription(ctx context.Context, userID *string) (*entity.Subscription, error) {
	sub, err := s.subscriptionsRepo.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return sub, nil
}

func (s *subscriptionService) CancelSubscription(ctx context.Context, userID *string) (*entity.Subscription, error) {
	// Retrieve the current subscription record from the database.
	subRecord, err := s.subscriptionsRepo.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Set your Stripe API key.
	stripe.Key = s.stripeSecretKey

	// Call the Stripe API to cancel the subscription.
	canceledSub, err := subscription.Cancel(subRecord.StripeSubscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel stripe subscription: %w", err)
	}

	// Update your subscription record. For example, set the status to canceled.
	subRecord.Status = string(canceledSub.Status)
	// Optionally record cancellation time (if your entity has such a field)
	subRecord.UpdatedAt = time.Now()

	updatedSub, err := s.subscriptionsRepo.Update(ctx, subRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription record: %w", err)
	}

	return updatedSub, nil
}

// SubscribeUser creates a new Stripe subscription for the given user and stores it.
func (s *subscriptionService) SubscribeUser(ctx context.Context, user *entity.User) (*entity.Subscription, error) {
	// Set the Stripe API key
	stripe.Key = s.stripeSecretKey

	var stripeCustomerID *string
	if user.StripeCustomerID.Valid {
		stripeCustomerID = &user.StripeCustomerID.String
	}

	// Define the subscription parameters
	params := &stripe.SubscriptionParams{
		Customer:        stripeCustomerID,
		Items:           []*stripe.SubscriptionItemsParams{{Price: stripe.String("price_1RCJhLGhK964Xz608A8iJO80")}},
		TrialPeriodDays: stripe.Int64(7),
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
