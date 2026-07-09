package commands

import (
	"context"
	"errors"
	"testing"
	"transactionhub/internal/domain/account"
	"transactionhub/internal/domain/transaction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//Testes Unitários:

// Testa se a transação original não é encontrada
func TestReversal_OriginalTransactionNotFound(t *testing.T) {
	mockAccountRepository := new(MockAccountRepository)
	mockTransactionRepository := new(MockTransactionRepository)

	handler := NewReversalHandler(mockAccountRepository, mockTransactionRepository)

	cmd := ReversalCommand{
		OriginalTransactionID: "tx-123",
		ReferenceID:           "ref-123",
	}

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "ref-123").
		Return((*transaction.Transaction)(nil), nil)

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "tx-123").
		Return((*transaction.Transaction)(nil), errors.New("not found")).Once()

	tx, err := handler.Execute(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, "original transaction not found", err.Error())
	assert.Nil(t, tx)

	mockAccountRepository.AssertExpectations(t)
	mockTransactionRepository.AssertExpectations(t)
}

// Testa se apenas transações bem sucedidas podem ser revertidas
func TestReversal_OnlySuccessfulTransactionsCanBeReversed(t *testing.T) {
	mockAccountRepository := new(MockAccountRepository)
	mockTransactionRepository := new(MockTransactionRepository)

	handler := NewReversalHandler(mockAccountRepository, mockTransactionRepository)

	cmd := ReversalCommand{
		OriginalTransactionID: "tx-123",
		ReferenceID:           "ref-123",
	}

	originalTx := &transaction.Transaction{
		ID:        "tx-123",
		AccountID: "account-123",
		Amount:    1000,
		Currency:  "USD",
		Operation: transaction.Credit,
		Status:    transaction.Failed,
	}

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "ref-123").
		Return((*transaction.Transaction)(nil), nil)

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "tx-123").
		Return(originalTx, nil)

	tx, err := handler.Execute(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, "only successful transactions can be reversed", err.Error())
	assert.NotNil(t, tx)
	assert.Equal(t, originalTx, tx)

	mockAccountRepository.AssertExpectations(t)
	mockTransactionRepository.AssertExpectations(t)
}

// Testa se o estorno de uma transação de crédito é concluído com sucesso
func TestReversal_Success_CreditOperation(t *testing.T) {
	mockAccountRepository := new(MockAccountRepository)
	mockTransactionRepository := new(MockTransactionRepository)

	handler := NewReversalHandler(mockAccountRepository, mockTransactionRepository)

	cmd := ReversalCommand{
		OriginalTransactionID: "tx-123",
		ReferenceID:           "ref-123",
	}

	originalTx := &transaction.Transaction{
		ID:        "tx-123",
		AccountID: "account-123",
		Amount:    1000,
		Currency:  "USD",
		Operation: transaction.Credit,
		Status:    transaction.Sucess,
	}

	acc := &account.Account{
		ID:      "account-123",
		Balance: 1000,
		Status:  "active",
	}

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "ref-123").
		Return((*transaction.Transaction)(nil), nil)

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "tx-123").
		Return(originalTx, nil)

	mockAccountRepository.On("FindByID", mock.Anything, "account-123").
		Return(acc, nil)

	mockAccountRepository.On("Save", mock.Anything, mock.Anything).
		Return(nil)

	mockTransactionRepository.On("Save", mock.Anything, mock.Anything).
		Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, transaction.Sucess, tx.Status)
	assert.Equal(t, transaction.Reversal, tx.Operation)
	assert.Equal(t, int64(0), acc.Balance)

	mockAccountRepository.AssertExpectations(t)
	mockTransactionRepository.AssertExpectations(t)
}
