package subscriptions

import (
	"context"
	"fmt"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/paymenttransactions"
	"github.com/google/uuid"
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
	HandleInvoiceSuccess(
		ctx context.Context,
		stripeSubID *string,
		amount *int64,
		currency string,
	) error
}

type subscriptionService struct {
	logger                    *zap.Logger
	stripeSecretKey           string
	subscriptionsRepo         storage.SubscriptionsRepository
	paymentTransactionService paymenttransactions.PaymentTransactionService
}

func NewSubscriptionService(
	logger *zap.Logger,
	stripeSecretKey string,
	subscriptionsRepo storage.SubscriptionsRepository,
	paymentTransactionService paymenttransactions.PaymentTransactionService,
) SubscriptionService {
	return &subscriptionService{
		logger:                    logger,
		stripeSecretKey:           stripeSecretKey,
		subscriptionsRepo:         subscriptionsRepo,
		paymentTransactionService: paymentTransactionService,
	}
}

func (s *subscriptionService) HandleInvoiceSuccess(
	ctx context.Context,
	stripeSubID *string,
	amount *int64,
	currency string,
) error {
	// 1. Find the subscription in the DB
	sub, err := s.subscriptionsRepo.GetSubscriptionByStripeID(ctx, stripeSubID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// 2. Create a payment transaction record
	payment := &entity.PaymentTransaction{
		ID:        uuid.NewString(),
		UserID:    sub.UserID,
		Amount:    int(*amount),
		Status:    "succeeded",
		Currency:  currency,
		CreatedAt: time.Now(),
	}

	if err := s.paymentTransactionService.InsertPaymentTransaction(ctx, payment); err != nil {
		return fmt.Errorf("failed to record payment: %w", err)
	}

	// 3. Update subscription status if still "trialing" or some other status
	if sub.Status != "active" {
		sub.Status = "active"
		sub.UpdatedAt = time.Now()

		if _, err := s.subscriptionsRepo.Update(ctx, sub); err != nil {
			return fmt.Errorf("failed to update subscription status: %w", err)
		}
	}

	return nil
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
