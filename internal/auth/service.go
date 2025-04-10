package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth/domain"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth/mappers"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth/storage"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
)

type UserService interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	RegisterNewUser(ctx context.Context, user *domain.User) (*entity.User, error)
	GenerateToken(user *entity.User) (string, error)
	UpdateUserDetails(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserById(ctx context.Context, id string) (*entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUserByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*entity.User, error)
}

type userService struct {
	logger          *zap.Logger
	userRepo        storage.UserRepository
	jwtSecret       []byte
	stripeSecretKey string
}

func NewUserService(
	logger *zap.Logger,
	userRepo storage.UserRepository,
	jwtSecret []byte,
	stripeApiKey string,
) UserService {
	return &userService{
		logger:          logger,
		userRepo:        userRepo,
		jwtSecret:       jwtSecret,
		stripeSecretKey: stripeApiKey,
	}
}

func (s *userService) GetUserByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*entity.User, error) {
	user, err := s.userRepo.GetUserByStripeCustomerID(ctx, stripeCustomerID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) RegisterNewUser(ctx context.Context, user *domain.User) (*entity.User, error) {
	s.logger.Sugar().Infof("Registering new user")

	stripe.Key = s.stripeSecretKey

	params := &stripe.CustomerParams{
		Email: &user.Email,
	}

	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe customer: %w", err)
	}

	userEntity := mappers.ToUserEntity(*user, cust)

	insertedUser, err := s.userRepo.InsertUser(ctx, userEntity)
	if err != nil {
		return nil, err
	}

	return insertedUser, nil
}

// GetUserByEmail retrieves a user by their email.
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	s.logger.Sugar().Infof("Getting user by email")

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GenerateToken generates a JWT for the authenticated user.
func (s *userService) GenerateToken(user *entity.User) (string, error) {
	s.logger.Sugar().Infof("Generating token")
	// Define claims; you can add custom claims as needed.
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // token expires in 24 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (s *userService) GetUserById(ctx context.Context, id string) (*entity.User, error) {
	s.logger.Sugar().Infof("Getting user by id")

	user, err := s.userRepo.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserDetails updates a user in the database.
func (s *userService) UpdateUserDetails(ctx context.Context, user *entity.User) (*entity.User, error) {
	s.logger.Sugar().Infof("Updating user details")

	user, err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user by their ID.
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	s.logger.Sugar().Infof("Deleting user")

	if err := s.userRepo.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("service error deleting user: %w", err)
	}

	return nil
}
