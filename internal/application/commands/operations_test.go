package commands

import (
	"context"
	"testing"
	"transactionhub/internal/domain/account"
	"transactionhub/internal/domain/transaction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---- Credit ----
// Testa falha ao creditar em uma conta inexistente
func TestCredit_AccountNotFound(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewCreditHandler(mockAR, mockTR)

	cmd := CreditCommand{AccountID: "acc-1", Amount: 500, Currency: "BRL", ReferenceID: "ref-1"}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-1").Return((*transaction.Transaction)(nil), nil)
	mockAR.On("FindByID", mock.Anything, "acc-1").Return((*account.Account)(nil), assert.AnError)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, true, "account not found")
	assert.Nil(t, tx)
	mockAR.AssertExpectations(t)
	mockTR.AssertExpectations(t)
}

// Testa sucesso ao realizar um crédito
func TestCredit_Success(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewCreditHandler(mockAR, mockTR)

	cmd := CreditCommand{AccountID: "acc-1", Amount: 1000, Currency: "BRL", ReferenceID: "ref-2"}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-2").Return((*transaction.Transaction)(nil), nil)
	mockAR.On("FindByID", mock.Anything, "acc-1").Return(&account.Account{ID: "acc-1", Balance: 0, Status: "active"}, nil)
	mockAR.On("Save", mock.Anything, mock.Anything).Return(nil)
	mockTR.On("Save", mock.Anything, mock.Anything).Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, false, "")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, transaction.Sucess, tx.Status)
	assert.Equal(t, transaction.Credit, tx.Operation)
	mockAR.AssertExpectations(t)
	mockTR.AssertExpectations(t)
}

// Testa idempotência para evitar crédito duplicado
func TestCredit_Idempotency(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewCreditHandler(mockAR, mockTR)

	cmd := CreditCommand{AccountID: "acc-1", Amount: 1000, Currency: "BRL", ReferenceID: "ref-dup"}
	existing := &transaction.Transaction{ID: "tx-existing", Status: transaction.Sucess}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-dup").Return(existing, nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, false, "")
	assert.NoError(t, err)
	assert.Equal(t, existing, tx)
	mockTR.AssertExpectations(t)
}

// ---- Debit ----
// Testa falha ao debitar com saldo insuficiente
func TestDebit_InsufficientBalance(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewDebitHandler(mockAR, mockTR)

	cmd := DebitCommand{AccountID: "acc-1", Amount: 9999, Currency: "BRL", ReferenceID: "ref-3"}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-3").Return((*transaction.Transaction)(nil), nil)
	mockAR.On("FindByID", mock.Anything, "acc-1").Return(&account.Account{ID: "acc-1", Balance: 100, Status: "active"}, nil)
	mockTR.On("Save", mock.Anything, mock.Anything).Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, true, "insufficient balance")
	assert.Error(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, transaction.Failed, tx.Status)
	mockAR.AssertExpectations(t)
	mockTR.AssertExpectations(t)
}

// Testa sucesso ao realizar um débito
func TestDebit_Success(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewDebitHandler(mockAR, mockTR)

	cmd := DebitCommand{AccountID: "acc-1", Amount: 500, Currency: "BRL", ReferenceID: "ref-4"}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-4").Return((*transaction.Transaction)(nil), nil)
	mockAR.On("FindByID", mock.Anything, "acc-1").Return(&account.Account{ID: "acc-1", Balance: 2000, Status: "active"}, nil)
	mockAR.On("Save", mock.Anything, mock.Anything).Return(nil)
	mockTR.On("Save", mock.Anything, mock.Anything).Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, false, "")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, transaction.Sucess, tx.Status)
	assert.Equal(t, transaction.Debit, tx.Operation)
	mockAR.AssertExpectations(t)
	mockTR.AssertExpectations(t)
}

// ---- Reserve ----
// Testa falha ao reservar saldo insuficiente
func TestReserve_InsufficientBalance(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewReserveHandler(mockAR, mockTR)

	cmd := ReserveCommand{AccountID: "acc-1", Amount: 5000, Currency: "BRL", ReferenceID: "ref-5"}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-5").Return((*transaction.Transaction)(nil), nil)
	mockAR.On("FindByID", mock.Anything, "acc-1").Return(&account.Account{ID: "acc-1", Balance: 100, Status: "active"}, nil)
	mockTR.On("Save", mock.Anything, mock.Anything).Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, true, "insufficient balance")
	assert.Error(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, transaction.Failed, tx.Status)
	mockAR.AssertExpectations(t)
	mockTR.AssertExpectations(t)
}

// Testa sucesso ao reservar saldo
func TestReserve_Success(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewReserveHandler(mockAR, mockTR)

	cmd := ReserveCommand{AccountID: "acc-1", Amount: 300, Currency: "BRL", ReferenceID: "ref-6"}
	acc := &account.Account{ID: "acc-1", Balance: 1000, Status: "active"}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-6").Return((*transaction.Transaction)(nil), nil)
	mockAR.On("FindByID", mock.Anything, "acc-1").Return(acc, nil)
	mockAR.On("Save", mock.Anything, mock.Anything).Return(nil)
	mockTR.On("Save", mock.Anything, mock.Anything).Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, false, "")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, transaction.Sucess, tx.Status)
	assert.Equal(t, transaction.Reserve, tx.Operation)
	assert.Equal(t, int64(300), acc.ReservedBalance)
	mockAR.AssertExpectations(t)
	mockTR.AssertExpectations(t)
}

// ---- Capture ----
// Testa falha ao capturar sem saldo reservado
func TestCapture_InsufficientReservedBalance(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewCaptureHandler(mockAR, mockTR)

	cmd := CaptureCommand{AccountID: "acc-1", Amount: 500, Currency: "BRL", ReferenceID: "ref-7"}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-7").Return((*transaction.Transaction)(nil), nil)
	// ReservedBalance=0, então Capture de 500 vai falhar
	mockAR.On("FindByID", mock.Anything, "acc-1").Return(&account.Account{ID: "acc-1", Balance: 1000, ReservedBalance: 0, Status: "active"}, nil)
	mockTR.On("Save", mock.Anything, mock.Anything).Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	assert.Error(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, transaction.Failed, tx.Status)
	mockAR.AssertExpectations(t)
	mockTR.AssertExpectations(t)
}

// Testa sucesso ao capturar saldo previamente reservado
func TestCapture_Success(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewCaptureHandler(mockAR, mockTR)

	cmd := CaptureCommand{AccountID: "acc-1", Amount: 200, Currency: "BRL", ReferenceID: "ref-8"}
	acc := &account.Account{ID: "acc-1", Balance: 1000, ReservedBalance: 200, Status: "active"}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-8").Return((*transaction.Transaction)(nil), nil)
	mockAR.On("FindByID", mock.Anything, "acc-1").Return(acc, nil)
	mockAR.On("Save", mock.Anything, mock.Anything).Return(nil)
	mockTR.On("Save", mock.Anything, mock.Anything).Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, false, "")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, transaction.Sucess, tx.Status)
	assert.Equal(t, transaction.Capture, tx.Operation)
	assert.Equal(t, int64(0), acc.ReservedBalance)
	assert.Equal(t, int64(800), acc.Balance)
	mockAR.AssertExpectations(t)
	mockTR.AssertExpectations(t)
}

// ---- Reversal de Transferência ----
// Testa sucesso ao estornar uma transferência
func TestReversal_Success_TransferOperation(t *testing.T) {
	mockAR := new(MockAccountRepository)
	mockTR := new(MockTransactionRepository)
	handler := NewReversalHandler(mockAR, mockTR)

	cmd := ReversalCommand{
		OriginalTransactionID: "tx-transfer",
		ReferenceID:           "ref-reversal",
	}

	originalTx := &transaction.Transaction{
		ID:        "tx-transfer",
		AccountID: "acc-source",
		Amount:    500,
		Currency:  "BRL",
		Operation: transaction.Transfer,
		Status:    transaction.Sucess,
		Metadata:  map[string]interface{}{"destination_account_id": "acc-dest"},
	}

	sourceAcc := &account.Account{ID: "acc-source", Balance: 0, Status: "active"}
	destAcc := &account.Account{ID: "acc-dest", Balance: 500, Status: "active"}

	mockTR.On("FindByReferenceID", mock.Anything, "ref-reversal").Return((*transaction.Transaction)(nil), nil)
	mockTR.On("FindByID", mock.Anything, "tx-transfer").Return(originalTx, nil)
	mockAR.On("FindByID", mock.Anything, "acc-source").Return(sourceAcc, nil)
	mockAR.On("FindByID", mock.Anything, "acc-dest").Return(destAcc, nil)
	mockAR.On("Save", mock.Anything, mock.Anything).Return(nil)
	mockTR.On("Save", mock.Anything, mock.Anything).Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, false, "")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, transaction.Sucess, tx.Status)
	assert.Equal(t, transaction.Reversal, tx.Operation)
	// Conta destino foi debitada
	assert.Equal(t, int64(0), destAcc.Balance)
	// Conta origem recebeu o valor de volta
	assert.Equal(t, int64(500), sourceAcc.Balance)
	mockAR.AssertExpectations(t)
	mockTR.AssertExpectations(t)
}
