package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/subscriptions"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/render"
)

type WebhookHandler interface {
	HandleStripeWebhook() http.HandlerFunc
}

// Handler is a structure that could hold dependencies like a logger and service interfaces.
type webhookHandler struct {
	logger              *zap.Logger
	stripeWebhookSecret string
	subscriptionService subscriptions.SubscriptionService
}

// NewHandler constructs a new webhook handler.
func NewWebhookHandler(
	logger *zap.Logger,
	stripeWebhookSecret string,
	subscriptionService subscriptions.SubscriptionService,
) WebhookHandler {
	return &webhookHandler{
		logger:              logger,
		stripeWebhookSecret: stripeWebhookSecret,
		subscriptionService: subscriptionService,
	}
}

// HandleStripeWebhook processes incoming Stripe webhook events.
func (h *webhookHandler) HandleStripeWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

		// Read the raw body.
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error("Error reading request body", zap.Error(err))
			http.Error(w, "Error reading request body", http.StatusServiceUnavailable)

			return
		}

		// Retrieve the Stripe-Signature header.
		sigHeader := r.Header.Get("Stripe-Signature")
		if sigHeader == "" {
			h.logger.Error("No Stripe-Signature header found")
			http.Error(w, "Missing Stripe-Signature header", http.StatusBadRequest)

			return
		}

		// Verify the event by constructing it from the raw payload and signature header.
		event, err := webhook.ConstructEvent(payload, sigHeader, h.stripeWebhookSecret)
		if err != nil {
			h.logger.Error("Failed to verify webhook signature", zap.Error(err))
			http.Error(w, fmt.Sprintf("Signature verification failed: %v", err), http.StatusBadRequest)

			return
		}

		// Optional: Log the event type for debugging.
		h.logger.Info("Received Stripe event", zap.String("event_type", string(event.Type)))

		// Handle the event based on its type.
		switch event.Type {
		case "invoice.paid":
			h.handleInvoicePaymentSucceeded(ctx, event)
		case "invoice.payment_failed":
			h.handleInvoicePaymentFailed(ctx, event)
		case "customer.subscription.updated":
			h.handleCustomerSubscriptionUpdated(ctx, event)
		case "customer.subscription.deleted":
			h.handleCustomerSubscriptionDeleted(ctx, event)

		default:
			h.logger.Info("Unhandled event type", zap.String("type", string(event.Type)))
		}

		// Respond with a 200 OK to acknowledge receipt of the event.
		render.Json(w, http.StatusOK, nil)
	}
}

func (h *webhookHandler) handleCustomerSubscriptionDeleted(ctx context.Context, event stripe.Event) {
	h.logger.Info("Handling customer.subscription.deleted")

	if err := h.subscriptionService.HandleSubscriptionDeleted(ctx, event); err != nil {
		h.logger.Error("Failed to handle subscription.deleted", zap.Error(err))
	}
}

// Stub function for handling successful invoice payment events.
func (h *webhookHandler) handleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event) {
	h.logger.Info("Handling invoice.payment_succeeded")

	// TODO: Implement your business logic here.
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		h.logger.Error("Error unmarshaling invoice event", zap.Error(err))
		return
	}

	// Extract required Stripe IDs
	stripeSubID := &invoice.Parent.SubscriptionDetails.Subscription.ID
	// Total amount paid (in cents)
	amount := &invoice.AmountPaid // in cents
	currency := invoice.Currency

	// Call your service layer to handle it
	err := h.subscriptionService.HandleInvoiceSuccess(
		ctx,
		stripeSubID,
		amount,
		string(currency),
	)
	if err != nil {
		h.logger.Error("Failed to process invoice.payment_succeeded", zap.Error(err))
	}
}

// Stub function for handling subscription updates.
func (h *webhookHandler) handleCustomerSubscriptionUpdated(ctx context.Context, event stripe.Event) {
	h.logger.Info("Handling customer.subscription.updated")

	err := h.subscriptionService.HandleSubscriptionUpdated(ctx, event)
	if err != nil {
		h.logger.Error("Failed to handle customer.subscription.updated", zap.Error(err))
	}
}

// Stub function for handling failed  payments.
func (h *webhookHandler) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) {
	h.logger.Info("Handling invoice.payment_failed")

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		h.logger.Error("Error unmarshaling invoice failed event", zap.Error(err))
		return
	}

	err := h.subscriptionService.HandleInvoiceFailed(
		ctx,
		invoice.Customer.ID,
		invoice.AmountDue,
		string(invoice.Currency),
	)
	if err != nil {
		h.logger.Error("Failed to handle payment failure", zap.Error(err))
	}
}
