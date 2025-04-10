package storage

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
)

type SubscriptionsRepository interface {
	Insert(ctx context.Context, subscription *entity.Subscription) (*entity.Subscription, error)
	GetSubscriptionByUserID(ctx context.Context, userID *string) (*entity.Subscription, error)
	Update(ctx context.Context, subscription *entity.Subscription) (*entity.Subscription, error)
}

type subscriptionsRepository struct {
	db *sqlx.DB
}

func NewSubscriptionsRepository(db *sqlx.DB) SubscriptionsRepository {
	return &subscriptionsRepository{
		db: db,
	}
}

func (s *subscriptionsRepository) Update(ctx context.Context, subscription *entity.Subscription) (*entity.Subscription, error) {
	_, err := subscription.Update(ctx, s.db, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return subscription, nil
}

func (s *subscriptionsRepository) GetSubscriptionByUserID(ctx context.Context, userID *string) (*entity.Subscription, error) {
	subscriptions, err := entity.Subscriptions(
		qm.Where("user_id = ?", *userID),
	).All(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription for user %s: %w", *userID, err)
	}

	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("no subscription found for user %s", *userID)
	}
	// Return the first subscription found.
	return subscriptions[0], nil
}

func (s *subscriptionsRepository) Insert(ctx context.Context, subscription *entity.Subscription) (*entity.Subscription, error) {
	err := subscription.Insert(ctx, s.db, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("failed to insert subscription into database: %s", err)
	}

	return subscription, nil
}
