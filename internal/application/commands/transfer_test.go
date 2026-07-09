package commands

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"transactionhub/internal/domain/account"
	"transactionhub/internal/domain/transaction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks:
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Save(ctx context.Context, account *account.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockAccountRepository) FindByID(ctx context.Context, id string) (*account.Account, error) {
	args := m.Called(ctx, id)

	if args.Get(0) != nil {
		return args.Get(0).(*account.Account), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Save(ctx context.Context, transaction *transaction.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByID(ctx context.Context, id string) (*transaction.Transaction, error) {
	args := m.Called(ctx, id)

	if args.Get(0) != nil {
		return args.Get(0).(*transaction.Transaction), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockTransactionRepository) FindByReferenceID(ctx context.Context, referenceID string) (*transaction.Transaction, error) {
	args := m.Called(ctx, referenceID)

	if args.Get(0) != nil {
		return args.Get(0).(*transaction.Transaction), args.Error(1)
	}

	return nil, args.Error(1)
}

// Função auxiliar que imprime PASS/FAILED na tela de acordo com o resultado do teste.
func checkResult(t *testing.T, err error, expectError bool, expectedErrMsg string) {
	t.Helper()

	switch {
	case expectError && err != nil && err.Error() == expectedErrMsg:
		fmt.Printf("[PASS] %s -> expected error received: %q\n", t.Name(), err.Error())

	case expectError && err != nil:
		fmt.Printf("[FAILED] %s -> expected error %q, but got %q\n", t.Name(), expectedErrMsg, err.Error())
		t.Fail()

	case expectError && err == nil:
		fmt.Printf("[FAILED] %s -> expected error %q, but no error occurred\n", t.Name(), expectedErrMsg)
		t.Fail()

	case !expectError && err == nil:
		fmt.Printf("[PASS] %s -> executed with no errors\n", t.Name())

	default:
		fmt.Printf("[FAILED] %s -> unexpected error: %q\n", t.Name(), err.Error())
		t.Fail()
	}
}

//Testes Unitários:

// Teste de transferencia para a mesma conta origem
func TestTransfer_SameAccount(t *testing.T) {

	mockAccountRepository := new(MockAccountRepository)
	mockTransactionRepository := new(MockTransactionRepository)
	handler := NewTransferHandler(mockAccountRepository, mockTransactionRepository)

	cmd := TransferCommand{
		SourceAccountID:      "account-123",
		DestinationAccountID: "account-123",
		Amount:               1000,
		Currency:             "USD",
		ReferenceID:          "ref-123",
	}

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "ref-123").
		Return((*transaction.Transaction)(nil), nil)

	mockAccountRepository.On("FindByID", mock.Anything, "account-123").
		Return(&account.Account{ID: "account-123", Status: "active"}, nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, true, "source and destination accounts cannot be the same")

	assert.Error(t, err)
	assert.Equal(t, "source and destination accounts cannot be the same", err.Error())
	assert.Nil(t, tx)

	mockAccountRepository.AssertExpectations(t)
	mockTransactionRepository.AssertExpectations(t)
}

// Testa se a conta destino existe
func TestTransfer_DestinationAccountNotFound(t *testing.T) {

	mockAccountRepository := new(MockAccountRepository)
	mockTransactionRepository := new(MockTransactionRepository)
	handler := NewTransferHandler(mockAccountRepository, mockTransactionRepository)

	cmd := TransferCommand{
		SourceAccountID:      "account-123",
		DestinationAccountID: "account-789",
		Amount:               1000,
		Currency:             "USD",
		ReferenceID:          "ref-123",
	}

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "ref-123").
		Return((*transaction.Transaction)(nil), nil)

	mockAccountRepository.On("FindByID", mock.Anything, "account-123").
		Return(&account.Account{ID: "account-123", Status: "active"}, nil)

	mockAccountRepository.On("FindByID", mock.Anything, "account-789").
		Return((*account.Account)(nil), errors.New("destination account not found")).Once()

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, true, "destination account not found")

	assert.Error(t, err)
	assert.Equal(t, "destination account not found", err.Error())
	assert.Nil(t, tx)

	mockAccountRepository.AssertExpectations(t)
	mockTransactionRepository.AssertExpectations(t)
}

// Testa se a conta possui saldo insuficiente
func TestTransfer_InsufficientBalance(t *testing.T) {
	mockAccountRepository := new(MockAccountRepository)
	mockTransactionRepository := new(MockTransactionRepository)

	handler := NewTransferHandler(mockAccountRepository, mockTransactionRepository)

	cmd := TransferCommand{
		SourceAccountID:      "account-123",
		DestinationAccountID: "account-456",
		Amount:               1000,
		Currency:             "USD",
		ReferenceID:          "ref-123",
	}

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "ref-123").
		Return((*transaction.Transaction)(nil), nil)

	mockAccountRepository.On("FindByID", mock.Anything, "account-123").
		Return(&account.Account{ID: "account-123", Balance: 100, Status: "active"}, nil)

	mockAccountRepository.On("FindByID", mock.Anything, "account-456").
		Return(&account.Account{ID: "account-456", Balance: 0, Status: "active"}, nil)

	mockTransactionRepository.On("Save", mock.Anything, mock.Anything).
		Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, true, "insufficient balance")

	assert.Error(t, err)
	assert.Equal(t, "insufficient balance", err.Error())
	assert.NotNil(t, tx)

	mockAccountRepository.AssertExpectations(t)
	mockTransactionRepository.AssertExpectations(t)
}

// Testa se a transferência foi concluida com sucesso
func TestTransfer_Success(t *testing.T) {
	mockAccountRepository := new(MockAccountRepository)
	mockTransactionRepository := new(MockTransactionRepository)

	handler := NewTransferHandler(mockAccountRepository, mockTransactionRepository)

	cmd := TransferCommand{
		SourceAccountID:      "account-123",
		DestinationAccountID: "account-456",
		Amount:               1000,
		Currency:             "USD",
		ReferenceID:          "ref-123",
	}

	mockTransactionRepository.On("FindByReferenceID", mock.Anything, "ref-123").
		Return((*transaction.Transaction)(nil), nil)

	mockAccountRepository.On("FindByID", mock.Anything, "account-123").
		Return(&account.Account{ID: "account-123", Balance: 2000, Status: "active"}, nil)

	mockAccountRepository.On("FindByID", mock.Anything, "account-456").
		Return(&account.Account{ID: "account-456", Balance: 0, Status: "active"}, nil)

	mockAccountRepository.On("Save", mock.Anything, mock.Anything).
		Return(nil)
	mockTransactionRepository.On("Save", mock.Anything, mock.Anything).
		Return(nil)

	tx, err := handler.Execute(context.Background(), cmd)

	checkResult(t, err, false, "")

	assert.NoError(t, err)
	assert.NotNil(t, tx)

	mockAccountRepository.AssertExpectations(t)
	mockTransactionRepository.AssertExpectations(t)
}
