package subscriptions

import (
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/subscriptions/dto"
	"net/http"

	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/subscriptions"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/context"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/render"
)

type SubscriptionsHandler interface {
	Subscribe() http.HandlerFunc
	Cancel() http.HandlerFunc
	Status() http.HandlerFunc
}

type subscriptionsHandler struct {
	logger               *zap.Logger
	subscriptionsService subscriptions.SubscriptionService
	userService          auth.UserService
}

func NewSubscriptionsHandler(
	logger *zap.Logger,
	subscriptionsService subscriptions.SubscriptionService,
	userService auth.UserService,
) SubscriptionsHandler {
	return &subscriptionsHandler{
		logger:               logger,
		subscriptionsService: subscriptionsService,
		userService:          userService,
	}
}

// Status returns the current subscription details.
func (h *subscriptionsHandler) Status() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get the user ID from the context
		userID, err := context.GetUserIDString(ctx)
		if err != nil {
			h.logger.Sugar().Errorw("user ID not found in session", "error", err)
			render.Json(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		// Use the subscription service to retrieve the user’s subscription record.
		subscription, err := h.subscriptionsService.GetUserSubscription(r.Context(), &userID)
		if err != nil {
			h.logger.Error("Failed to get subscription status", zap.Error(err))
			render.Json(w, http.StatusInternalServerError, "Unable to retrieve subscription status")

			return
		}

		resp, err := dto.ToStatusResponse(subscription)
		if err != nil {
			h.logger.Error("Failed to map subscription to status response", zap.Error(err))
			render.Json(w, http.StatusInternalServerError, "Unable to map subscription to status response")
		}

		// Return the subscription details as JSON.
		render.Json(w, http.StatusOK, resp)
	}
}

func (h *subscriptionsHandler) Cancel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, err := context.GetUserIDString(ctx)
		if err != nil {
			h.logger.Sugar().Errorw("user ID not found in session", "error", err)
			render.Json(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		// Call the service method to cancel the subscription.
		updatedSubscription, err := h.subscriptionsService.CancelSubscription(r.Context(), &userID)
		if err != nil {
			h.logger.Error("Failed to cancel subscription", zap.Error(err))
			render.Json(w, http.StatusInternalServerError, "Unable to cancel subscription")

			return
		}

		// Return a confirmation message along with subscription details.
		response := map[string]interface{}{
			"message":      "Subscription successfully cancelled",
			"subscription": updatedSubscription,
		}
		render.Json(w, http.StatusOK, response)
	}
}

func (h *subscriptionsHandler) Subscribe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get the user ID from the context
		userID, err := context.GetUserIDString(ctx)
		if err != nil {
			h.logger.Sugar().Errorw("user ID not found in session", "error", err)
			render.Json(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		userEntity, err := h.userService.GetUserById(ctx, userID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to get user by ID", "error", err)
			render.Json(w, http.StatusInternalServerError, "internal server error")

			return
		}

		subscription, err := h.subscriptionsService.SubscribeUser(ctx, userEntity)
		if err != nil {
			h.logger.Sugar().Errorw("failed to subscribe to user", "error", err)
			render.Json(w, http.StatusInternalServerError, "Failed to create subscription")

			return
		}

		// Return subscription status and billing info as JSON.
		response := map[string]interface{}{
			"subscriptionID": subscription.StripeSubscriptionID,
			"status":         subscription.Status,
			"trialStart":     subscription.TrialStart,
			"trialEnd":       subscription.TrialEnd,
		}

		render.Json(w, http.StatusOK, response)
	}
}
