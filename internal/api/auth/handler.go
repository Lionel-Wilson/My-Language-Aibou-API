package auth

import (
	"database/sql"
	"net/http"

	"github.com/friendsofgo/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/auth/dto"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth/domain"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/subscriptions"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/context"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/render"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/request"
)

type AuthHandler interface {
	Register() http.HandlerFunc
	Login() http.HandlerFunc
	UpdateDetails() http.HandlerFunc
	Delete() http.HandlerFunc
}

type handler struct {
	logger               *zap.Logger
	userService          auth.UserService
	subscriptionsService subscriptions.SubscriptionService
}

func NewAuthHandler(
	logger *zap.Logger,
	userService auth.UserService,
	subscriptionsService subscriptions.SubscriptionService,
) AuthHandler {
	return &handler{
		logger:               logger,
		userService:          userService,
		subscriptionsService: subscriptionsService,
	}
}

func (h *handler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.RegisterRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate register request body",
				"context", ctx, "error", err)

			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		userDomain, err := domain.RegisterRequestToUserDomain(req)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map register request to user domain",
				"context", ctx, "error", err)
			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		user, err := h.userService.RegisterNewUser(ctx, userDomain)
		if err != nil {
			h.logger.Sugar().Errorw("Error registering new user", "error", err)

			render.Json(w, http.StatusInternalServerError, err.Error())

			return
		}

		subscription, err := h.subscriptionsService.SubscribeUser(ctx, user)
		if err != nil {
			h.logger.Sugar().Errorw("failed to subscribe to user", "error", err)
			render.Json(w, http.StatusInternalServerError, "Failed to create subscription")

			return
		}

		// Generate an authentication token (e.g., JWT)
		token, err := h.userService.GenerateToken(user)
		if err != nil {
			h.logger.Sugar().Errorw("failed to generate token", "error", err)
			render.Json(w, http.StatusInternalServerError, "internal server error")

			return
		}

		resp := dto.ToRegisterResponse(user, subscription)

		render.Json(w, http.StatusCreated, map[string]interface{}{
			"token":       token,
			"userDetails": resp,
		})
	}
}

func (h *handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.LoginRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate login request body", "error", err)
			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		userEntity, err := h.userService.GetUserByEmail(ctx, req.Email)
		if err != nil {
			h.logger.Sugar().Errorw("failed to find user", "error", err)
			// For security, use the same error for "not found" or "wrong password"
			render.Json(w, http.StatusUnauthorized, "invalid credentials")

			return
		}

		// Compare the provided password with the stored hashed password.
		err = bcrypt.CompareHashAndPassword([]byte(userEntity.PasswordHash), []byte(req.Password))
		if err != nil {
			h.logger.Sugar().Errorw("password mismatch", "error", err)
			render.Json(w, http.StatusUnauthorized, "invalid credentials")

			return
		}

		// Generate an authentication token (e.g., JWT)
		token, err := h.userService.GenerateToken(userEntity)
		if err != nil {
			h.logger.Sugar().Errorw("failed to generate token", "error", err)
			render.Json(w, http.StatusInternalServerError, "internal server error")

			return
		}

		subscriptionEntity, err := h.subscriptionsService.GetUserSubscription(ctx, &userEntity.ID)
		if err != nil {
			// If no subscription exists, create one
			if errors.Is(err, sql.ErrNoRows) {
				subscriptionEntity, err = h.subscriptionsService.SubscribeUser(ctx, userEntity)
				if err != nil {
					h.logger.Sugar().Errorw("failed to subscribe user during login", "error", err)
					render.Json(w, http.StatusInternalServerError, "failed to subscribe user")

					return
				}
			} else {
				h.logger.Sugar().Errorw("failed to get subscription", "error", err)
				render.Json(w, http.StatusInternalServerError, "internal server error")

				return
			}
		}

		resp := dto.ToUserDetailsResponse(userEntity, subscriptionEntity)

		// Step 2e: Return the token in the response.
		render.Json(w, http.StatusOK, map[string]interface{}{
			"token":       token,
			"userDetails": resp,
		})
	}
}

func (h *handler) UpdateDetails() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Decode and validate the update request.
		var req dto.UpdateDetailsRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Errorw("failed to decode update details request", "error", err)
			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		// Get the user ID from the context
		userID, err := context.GetUserIDString(ctx)
		if err != nil {
			h.logger.Sugar().Errorw("user ID not found in session", "error", err)
			render.Json(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		// Retrieve the current user from the database.
		currentUser, err := h.userService.GetUserById(ctx, userID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to retrieve user", "error", err)
			render.Json(w, http.StatusInternalServerError, "internal server error")

			return
		}

		// Update user fields based on the request.
		h.updateUserFields(currentUser, req)

		// Call the service to update the user details.
		updatedUserDetails, err := h.userService.UpdateUserDetails(ctx, currentUser)
		if err != nil {
			h.logger.Sugar().Errorw("failed to update user details", "error", err)
			render.Json(w, http.StatusInternalServerError, "failed to update user")

			return
		}

		subscriptionEntity, err := h.subscriptionsService.GetUserSubscription(ctx, &updatedUserDetails.ID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to get subscription", "error", err)
			render.Json(w, http.StatusInternalServerError, "internal server error")
		}

		resp := dto.ToUserDetailsResponse(updatedUserDetails, subscriptionEntity)

		render.Json(w, http.StatusOK, map[string]interface{}{
			"message":     "user details updated successfully",
			"userDetails": resp,
		})
	}
}

func (h *handler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, err := context.GetUserIDString(ctx)
		if err != nil {
			h.logger.Sugar().Errorw("user ID not found in session", "error", err)
			render.Json(w, http.StatusUnauthorized, "unauthorized")

			return
		}

		// Call the service to delete the user.
		if err := h.userService.DeleteUser(ctx, userID); err != nil {
			h.logger.Sugar().Errorw("failed to delete user", "error", err)
			render.Json(w, http.StatusInternalServerError, "failed to delete user")

			return
		}

		// Return a success response.
		render.Json(w, http.StatusOK, map[string]string{"message": "user deleted successfully"})
	}
}

// updateUserFields updates the provided user entity with the values from the update request.
func (h *handler) updateUserFields(user *entity.User, req dto.UpdateDetailsRequest) {
	if req.Email != "" {
		user.Email = req.Email
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			h.logger.Sugar().Errorw("failed to hash new password", "error", err)
		} else {
			user.PasswordHash = string(hashedPassword)
		}
	}
}
