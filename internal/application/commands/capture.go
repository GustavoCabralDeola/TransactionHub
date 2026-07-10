package commands

import (
	"context"
	"errors"
	"transactionhub/internal/domain/account"
	"transactionhub/internal/domain/transaction"

	"github.com/google/uuid"
)

type CaptureCommand struct {
	AccountID   string
	Amount      int64
	Currency    string
	ReferenceID string
}

type CaptureHandler struct {
	accountRepo     account.Repository
	transactionRepo transaction.Repository
}

func NewCaptureHandler(ar account.Repository, tr transaction.Repository) *CaptureHandler {
	return &CaptureHandler{
		accountRepo:     ar,
		transactionRepo: tr,
	}
}

func (h *CaptureHandler) Execute(ctx context.Context, cmd CaptureCommand) (*transaction.Transaction, error) {
	defer LockAccount(cmd.AccountID)()

	existing, _ := h.transactionRepo.FindByReferenceID(ctx, cmd.ReferenceID)
	if existing != nil {
		return existing, nil
	}

	acc, err := h.accountRepo.FindByID(ctx, cmd.AccountID)
	if err != nil {
		return nil, errors.New("account not found")
	}

	tx := transaction.NewTransaction(
		uuid.New().String(),
		acc.ID,
		transaction.Capture,
		cmd.Amount,
		cmd.Currency,
		cmd.ReferenceID,
		nil,
	)

	if err := acc.Capture(cmd.Amount); err != nil {
		tx.MarkAsFailed(err.Error())
		_ = h.transactionRepo.Save(ctx, tx)
		return tx, err
	}

	tx.MarkAsSucess()
	_ = h.accountRepo.Save(ctx, acc)
	_ = h.transactionRepo.Save(ctx, tx)

	return tx, nil
}
