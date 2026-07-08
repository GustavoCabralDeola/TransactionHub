package commands

import (
	"context"
	"errors"
	"transactionhub/internal/domain/account"
	"transactionhub/internal/domain/transaction"

	"github.com/google/uuid"
)

type TransferCommand struct {
	SourceAccountID      string
	DestinationAccountID string
	Amount               int64
	Currency             string
	ReferenceID          string
}

type TransferHandler struct {
	accountRepo     account.Repository
	transactionRepo transaction.Repository
}

func NewTransferHandler(ar account.Repository, tr transaction.Repository) *TransferHandler {
	return &TransferHandler{
		accountRepo:     ar,
		transactionRepo: tr,
	}
}

// Faz a orquestração da transferência
func (h *TransferHandler) Execute(ctx context.Context, cmd TransferCommand) (*transaction.Transaction, error) {

	existingTx, _ := h.transactionRepo.FindByReferenceID(ctx, cmd.SourceAccountID)
	if existingTx != nil {
		return existingTx, nil
	}
	sourceAccount, err := h.accountRepo.FindByID(ctx, cmd.SourceAccountID)

	if err != nil {
		return nil, errors.New("failed to find source account")
	}

	if cmd.SourceAccountID == cmd.DestinationAccountID {
		return nil, errors.New("source and destination accounts cannot be the same")
	}

	destinationAccount, err := h.accountRepo.FindByID(ctx, cmd.DestinationAccountID)

	if err != nil {
		return nil, errors.New("destination account not found")
	}

	metadata := map[string]interface{}{

		"destination_account_id": destinationAccount.ID,
	}

	tx := transaction.NewTransaction(
		uuid.New().String(),
		sourceAccount.ID,
		transaction.Transfer,
		cmd.Amount,
		cmd.Currency,
		cmd.ReferenceID,
		metadata,
	)

	errDebit := sourceAccount.Debit(cmd.Amount)
	if errDebit != nil {
		tx.MarkAsFailed(errDebit.Error())
		_ = h.transactionRepo.Save(ctx, tx)

		return tx, errDebit
	}

	errCredit := destinationAccount.Credit(cmd.Amount)
	if errCredit != nil {
		_ = sourceAccount.Credit(cmd.Amount)

		tx.MarkAsFailed(errCredit.Error())
		_ = h.transactionRepo.Save(ctx, tx)
		return tx, errCredit
	}

	tx.MarkAsSucess()

	_ = h.accountRepo.Save(ctx, sourceAccount)
	_ = h.accountRepo.Save(ctx, destinationAccount)
	_ = h.transactionRepo.Save(ctx, tx)

	return tx, nil
}
