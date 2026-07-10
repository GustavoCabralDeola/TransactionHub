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

// NewReversalHandler cria um handler de estorno com as dependências injetadas.
func NewReversalHandler(ar account.Repository, tr transaction.Repository) *ReversalHandler {
	return &ReversalHandler{
		accountRepo:     ar,
		transactionRepo: tr,
	}
}

// Execute estorna uma transação anterior, reverte crédito com débito e vice-versa e garante idempotência via reference_id
func (h *ReversalHandler) Execute(ctx context.Context, cmd ReversalCommand) (*transaction.Transaction, error) {

	existingTx, _ := h.transactionRepo.FindByReferenceID(ctx, cmd.ReferenceID)
	if existingTx != nil {
		return existingTx, nil
	}

	originalTx, err := h.transactionRepo.FindByID(ctx, cmd.OriginalTransactionID)
	if err != nil {
		return nil, errors.New("original transaction not found")
	}

	if originalTx.Status != transaction.Sucess {
		return originalTx, errors.New("only successful transactions can be reversed")
	}

	if originalTx.Operation == transaction.Transfer {
		destID, ok := originalTx.Metadata["destination_account_id"].(string)
		if !ok || destID == "" {
			return nil, errors.New("could not determine destination account for transfer reversal")
		}
		defer LockTransferAccounts(originalTx.AccountID, destID)()
	} else {
		defer LockAccount(originalTx.AccountID)()
	}

	tx := transaction.NewTransaction(
		uuid.New().String(),
		originalTx.AccountID,
		transaction.Reversal,
		originalTx.Amount,
		originalTx.Currency,
		cmd.ReferenceID,
		map[string]interface{}{"original_transaction_id": originalTx.ID},
	)

	switch originalTx.Operation {
	case transaction.Credit:

		acc, err := h.accountRepo.FindByID(ctx, originalTx.AccountID)
		if err != nil {
			return nil, errors.New("account not found")
		}
		if err := acc.Debit(originalTx.Amount); err != nil {
			tx.MarkAsFailed(err.Error())
			_ = h.transactionRepo.Save(ctx, tx)
			return tx, err
		}
		tx.MarkAsSucess()
		_ = h.accountRepo.Save(ctx, acc)

	case transaction.Debit:

		acc, err := h.accountRepo.FindByID(ctx, originalTx.AccountID)
		if err != nil {
			return nil, errors.New("account not found")
		}
		if err := acc.Credit(originalTx.Amount); err != nil {
			tx.MarkAsFailed(err.Error())
			_ = h.transactionRepo.Save(ctx, tx)
			return tx, err
		}
		tx.MarkAsSucess()
		_ = h.accountRepo.Save(ctx, acc)

	case transaction.Transfer:

		destID := originalTx.Metadata["destination_account_id"].(string)

		sourceAcc, err := h.accountRepo.FindByID(ctx, originalTx.AccountID)
		if err != nil {
			return nil, errors.New("source account not found")
		}
		destAcc, err := h.accountRepo.FindByID(ctx, destID)
		if err != nil {
			return nil, errors.New("destination account not found")
		}

		if err := destAcc.Debit(originalTx.Amount); err != nil {
			tx.MarkAsFailed(err.Error())
			_ = h.transactionRepo.Save(ctx, tx)
			return tx, err
		}

		if err := sourceAcc.Credit(originalTx.Amount); err != nil {
			tx.MarkAsFailed(err.Error())
			_ = h.transactionRepo.Save(ctx, tx)
			return tx, err
		}

		tx.MarkAsSucess()
		_ = h.accountRepo.Save(ctx, sourceAcc)
		_ = h.accountRepo.Save(ctx, destAcc)

	default:
		return nil, errors.New("cannot reverse this operation type")
	}

	_ = h.transactionRepo.Save(ctx, tx)
	return tx, nil
}
