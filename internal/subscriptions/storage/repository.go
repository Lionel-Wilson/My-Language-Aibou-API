package storage

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
)

type SubscriptionsRepository interface {
	Insert(ctx context.Context, subscription *entity.Subscription) (*entity.Subscription, error)
}

type subscriptionsRepository struct {
	db *sqlx.DB
}

func NewSubscriptionsRepository(db *sqlx.DB) SubscriptionsRepository {
	return &subscriptionsRepository{
		db: db,
	}
}

func (s *subscriptionsRepository) Insert(ctx context.Context, subscription *entity.Subscription) (*entity.Subscription, error) {
	err := subscription.Insert(ctx, s.db, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("failed to insert subscription into database: %s", err)
	}

	return subscription, nil
}
