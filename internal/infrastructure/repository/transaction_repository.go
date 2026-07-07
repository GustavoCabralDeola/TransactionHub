package repository

import (
	"context"
	"encoding/json"
	"errors"
	"transactionhub/internal/domain/transaction"

	"gorm.io/gorm"
)

type GormTransactionRepository struct {
	db *gorm.DB
}

func NewGormTransactionRepository(db *gorm.DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db}
}

type TransactionModel struct {
	ID           string `gorm: "primary_key; column:id"`
	AccountID    string `gorm: "column:account_id"`
	Amount       int64  `gorm: "column:amount"`
	Currency     string `gorm: "column:currency"`
	ReferenceID  string `gorm: "column:reference_id"`
	Status       string `gorm: "column:status"`
	Metadata     []byte `gorm: "column:metadata"`
	ErrorMessage string `gorm: "column:error_message"`
}

func (TransactionModel) TableName() string {
	return "transactions"
}

func (r *GormTransactionRepository) Save(ctx context.Context, tx *transaction.Transaction) error {
	metaBytes, _ := json.Marshal(tx.Metadata)

	var errorMessage string
	if tx.ErrorMessage != nil {
		errorMessage = *tx.ErrorMessage
	}

	transactionModel := &TransactionModel{
		ID:           tx.ID,
		AccountID:    tx.AccountID,
		Amount:       tx.Amount,
		Currency:     tx.Currency,
		ReferenceID:  tx.ReferenceID,
		Status:       string(tx.Status),
		Metadata:     metaBytes,
		ErrorMessage: errorMessage,
	}

	return r.db.WithContext(ctx).Save(&transactionModel).Error
}

func (r *GormTransactionRepository) FindReferenceByID(ctx context.Context, referenceID string) (*transaction.Transaction, error) {

	var transactionModel TransactionModel

	err := r.db.WithContext(ctx).Where("reference_id = ?", referenceID).First(&transactionModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.mapToDomain(transactionModel), nil
}

//Auxiliar

func (r *GormTransactionRepository) mapToDomain(transactionModel TransactionModel) *transaction.Transaction {
	var metaData map[string]interface{}
	if len(transactionModel.Metadata) > 0 {
		_ = json.Unmarshal(transactionModel.Metadata, &metaData)
	}

	var errMsgPtr *string
	if transactionModel.ErrorMessage != "" {

		msg := transactionModel.ErrorMessage
		errMsgPtr = &msg
	}

	return &transaction.Transaction{
		ID:           transactionModel.ID,
		AccountID:    transactionModel.AccountID,
		Amount:       transactionModel.Amount,
		Currency:     transactionModel.Currency,
		ReferenceID:  transactionModel.ReferenceID,
		Status:       transaction.Status(transactionModel.Status),
		Metadata:     metaData,
		ErrorMessage: errMsgPtr,
	}
}
