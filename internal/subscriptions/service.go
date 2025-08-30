package subscriptions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/volatiletech/null/v8"
	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/paymenttransactions"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/subscriptions/storage"
)

//todo:don't return entity from service. convert to domain object
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
	HandleInvoiceFailed(
		ctx context.Context,
		stripeCustomerID string,
		amount int64,
		currency string,
	) error
	HandleSubscriptionUpdated(ctx context.Context, event stripe.Event) error
	HandleSubscriptionDeleted(ctx context.Context, event stripe.Event) error
	CreateCheckoutSession(ctx context.Context, userID string) (*stripe.CheckoutSession, error)
}

type subscriptionService struct {
	logger                    *zap.Logger
	stripeSecretKey           string
	subscriptionsRepo         storage.SubscriptionsRepository
	paymentTransactionService paymenttransactions.PaymentTransactionService
	userService               auth.UserService
	stripePriceID             string
	checkoutSuccessURL        string
	checkoutCancelURL         string
}

func NewSubscriptionService(
	logger *zap.Logger,
	stripeSecretKey string,
	subscriptionsRepo storage.SubscriptionsRepository,
	paymentTransactionService paymenttransactions.PaymentTransactionService,
	userService auth.UserService,
	stripePriceID string,
	checkoutSuccessURL string,
	checkoutCancelURL string,
) SubscriptionService {
	return &subscriptionService{
		logger:                    logger,
		stripeSecretKey:           stripeSecretKey,
		subscriptionsRepo:         subscriptionsRepo,
		paymentTransactionService: paymentTransactionService,
		userService:               userService,
		stripePriceID:             stripePriceID,
		checkoutSuccessURL:        checkoutSuccessURL,
		checkoutCancelURL:         checkoutCancelURL,
	}
}

func (s *subscriptionService) CreateCheckoutSession(ctx context.Context, userID string) (*stripe.CheckoutSession, error) {
	user, err := s.userService.GetUserById(ctx, userID)
	if err != nil {
		return nil, err
	}

	subscriptionEntity, err := s.subscriptionsRepo.GetSubscriptionByUserID(ctx, &userID)
	if err != nil {
		return nil, err
	}

	var sub entity.Subscription
	if subscriptionEntity != nil {
		sub = *subscriptionEntity
	}

	if sub.Status == "active" || sub.Status == "trialing" {
		return nil, fmt.Errorf("user already has an active or trialing subscription")
	}

	var trialEnd time.Time
	if sub.TrialEnd.Valid {
		trialEnd = sub.TrialEnd.Time
	}

	var stripeCustomerID string
	if user.StripeCustomerID.Valid {
		stripeCustomerID = user.StripeCustomerID.String
	}

	stripe.Key = s.stripeSecretKey

	if s.checkoutSuccessURL == "" || s.checkoutCancelURL == "" {
		return nil, fmt.Errorf("checkout URLs are not set")
	}

	sess, err := session.New(&stripe.CheckoutSessionParams{
		Customer: stripe.String(stripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(s.stripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialEnd: stripe.Int64(trialEnd.Unix()),
		},
		SuccessURL: stripe.String(s.checkoutSuccessURL),
		CancelURL:  stripe.String(s.checkoutCancelURL),
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (s *subscriptionService) HandleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	sub, err := s.subscriptionsRepo.GetSubscriptionByStripeID(ctx, &stripeSub.ID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	sub.Status = "canceled"
	sub.UpdatedAt = time.Now()

	if _, err := s.subscriptionsRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update canceled subscription: %w", err)
	}

	// Optional: clean up related resources

	return nil
}

func (s *subscriptionService) HandleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return fmt.Errorf("failed to parse subscription object: %w", err)
	}

	// 1. Get the subscription from your DB using the Stripe Subscription ID
	sub, err := s.subscriptionsRepo.GetSubscriptionByStripeID(ctx, &stripeSub.ID)
	if err != nil {
		return fmt.Errorf("could not find local subscription: %w", err)
	}

	// 2. Update the local subscription
	sub.Status = string(stripeSub.Status)
	sub.TrialStart = null.TimeFrom(time.Unix(stripeSub.TrialStart, 0))
	sub.TrialEnd = null.TimeFrom(time.Unix(stripeSub.TrialEnd, 0))
	sub.UpdatedAt = time.Now()

	if _, err := s.subscriptionsRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription in DB: %w", err)
	}

	return nil
}

func (s *subscriptionService) HandleInvoiceFailed(
	ctx context.Context,
	stripeCustomerID string,
	amount int64,
	currency string,
) error {
	// 1. Get user from customer ID
	user, err := s.userService.GetUserByStripeCustomerID(ctx, stripeCustomerID)
	if err != nil {
		return fmt.Errorf("user not found for stripe customer: %w", err)
	}

	// 2. Get subscription
	sub, err := s.subscriptionsRepo.GetSubscriptionByUserID(ctx, &user.ID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// 3. Update subscription status if desired
	sub.Status = "past_due"
	sub.UpdatedAt = time.Now()

	_, err = s.subscriptionsRepo.Update(ctx, sub)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// 4. Record a failed payment transaction
	tx := &entity.PaymentTransaction{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		Amount:    int(amount),
		Currency:  currency,
		Status:    "failed",
		CreatedAt: time.Now(),
	}
	if err := s.paymentTransactionService.InsertPaymentTransaction(ctx, tx); err != nil {
		return fmt.Errorf("failed to record failed payment: %w", err)
	}

	// 5. Optionally send email/notification (pseudo-code)
	// s.emailService.SendPaymentFailureNotification(user.Email)

	return nil
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
		Items:           []*stripe.SubscriptionItemsParams{{Price: stripe.String(s.stripePriceID)}},
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
