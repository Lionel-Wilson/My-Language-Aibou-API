package storage

import (
	"context"
	"fmt"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type PaymentTransactionRepository interface {
	Insert(ctx context.Context, paymentTransaction *entity.PaymentTransaction) error
}

type paymentTransactionRepository struct {
	db *sqlx.DB
}

func NewPaymentTransactionRepository(db *sqlx.DB) PaymentTransactionRepository {
	return &paymentTransactionRepository{
		db: db,
	}
}

func (s *paymentTransactionRepository) Insert(
	ctx context.Context,
	paymentTransaction *entity.PaymentTransaction,
) error {

	err := paymentTransaction.Insert(ctx, s.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to insert payment transaction into database: %s", err)
	}

	return nil
}
