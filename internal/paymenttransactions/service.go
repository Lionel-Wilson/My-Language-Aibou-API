package paymenttransactions

import (
	"context"

	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
	storage2 "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/paymenttransactions/storage"
)

type PaymentTransactionService interface {
	InsertPaymentTransaction(ctx context.Context, paymentTransaction *entity.PaymentTransaction) error
}

type paymentTransactionService struct {
	logger                 *zap.Logger
	paymentTransactionRepo storage2.PaymentTransactionRepository
}

func NewPaymentTransactionService(
	logger *zap.Logger,
	paymentTransactionRepo storage2.PaymentTransactionRepository,
) PaymentTransactionService {
	return &paymentTransactionService{
		logger:                 logger,
		paymentTransactionRepo: paymentTransactionRepo,
	}
}

func (s *paymentTransactionService) InsertPaymentTransaction(
	ctx context.Context,
	paymentTransaction *entity.PaymentTransaction,
) error {
	err := s.paymentTransactionRepo.Insert(ctx, paymentTransaction)
	if err != nil {
		return err
	}

	return nil
}
