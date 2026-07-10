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

// NewTransferHandler cria um handler de transferência com as dependências injetadas.
func NewTransferHandler(ar account.Repository, tr transaction.Repository) *TransferHandler {
	return &TransferHandler{
		accountRepo:     ar,
		transactionRepo: tr,
	}
}

// Realiza a transferência entre duas contas
func (h *TransferHandler) Execute(ctx context.Context, cmd TransferCommand) (*transaction.Transaction, error) {

	existingTx, _ := h.transactionRepo.FindByReferenceID(ctx, cmd.ReferenceID)
	if existingTx != nil {
		return existingTx, nil
	}

	if cmd.SourceAccountID == cmd.DestinationAccountID {
		return nil, errors.New("source and destination accounts cannot be the same")
	}

	defer LockTransferAccounts(cmd.SourceAccountID, cmd.DestinationAccountID)()

	sourceAccount, err := h.accountRepo.FindByID(ctx, cmd.SourceAccountID)
	if err != nil {
		return nil, errors.New("failed to find source account")
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

	if err := h.accountRepo.Save(ctx, sourceAccount); err != nil {
		return nil, errors.New("could not save source account: " + err.Error())
	}
	if err := h.accountRepo.Save(ctx, destinationAccount); err != nil {
		return nil, errors.New("could not save destination account: " + err.Error())
	}
	_ = h.transactionRepo.Save(ctx, tx)

	return tx, nil
}
