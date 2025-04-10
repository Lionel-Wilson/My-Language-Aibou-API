package subscriptions

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/subscriptions"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/context"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/render"
)

type SubscriptionsHandler interface {
	Subscribe() http.HandlerFunc
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
			"subscription_id": subscription.StripeSubscriptionID,
			"status":          subscription.Status,
			"trial_start":     subscription.TrialStart,
			"trial_end":       subscription.TrialEnd,
		}

		render.Json(w, http.StatusOK, response)
	}
}
