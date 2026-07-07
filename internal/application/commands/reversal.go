package commands

import (
	"context"
	"errors"
	"transactionhub/internal/domain/account"
	"transactionhub/internal/domain/transaction"

	"github.com/google/uuid"
)

type ReversalCommand struct {
	OriginalTransactionID string
	ReferenceID           string
}

type ReversalHandler struct {
	accountRepo     account.Repository
	transactionRepo transaction.Repository
}

func NewReversalHandler(ar account.Repository, tr transaction.Repository) *ReversalHandler {

	return &ReversalHandler{
		accountRepo:     ar,
		transactionRepo: tr,
	}
}

func (h *ReversalHandler) Execute(ctx context.Context, cmd ReversalCommand) (*transaction.Transaction, error) {
	existingTx, _ := h.transactionRepo.FindByReferenceID(ctx, cmd.ReferenceID)
	if existingTx != nil {
		return existingTx, nil
	}

	originalTx, err := h.transactionRepo.FindByReferenceID(ctx, cmd.OriginalTransactionID)
	if err != nil {
		return nil, errors.New("original transaction not found")
	}

	if originalTx.Status != transaction.Sucess {

		return originalTx, errors.New("only successful transactions can be reversed")
	}

	acc, err := h.accountRepo.FindByID(ctx, originalTx.AccountID)
	if err != nil {
		return nil, errors.New("account not found")
	}

	tx := transaction.NewTransaction(
		uuid.New().String(),
		acc.ID,
		transaction.Reversal,
		originalTx.Amount,
		originalTx.Currency,
		cmd.ReferenceID,
		map[string]interface{}{"original_transaction_id": originalTx.ID},
	)

	if originalTx.Operation == transaction.Credit {
		err = acc.Debit(originalTx.Amount)

	} else if originalTx.Operation == transaction.Debit {
		err = acc.Credit(originalTx.Amount)
	} else {
		return nil, errors.New("cannot reverse this operation type")
	}

	if err != nil {
		tx.MarkAsFailed(err.Error())
		_ = h.transactionRepo.Save(ctx, tx)

		return tx, err

	}

	tx.MarkAsSucess()

	_ = h.accountRepo.Save(ctx, acc)
	err = h.transactionRepo.Save(ctx, tx)

	return tx, nil
}
