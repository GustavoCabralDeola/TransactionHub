package repository

import (
	"context"
	"transactionhub/internal/domain/account"

	"gorm.io/gorm"
)

type GormAccountRepository struct {
	db *gorm.DB
}

func NewGormAccountRepository(db *gorm.DB) *GormAccountRepository {
	return &GormAccountRepository{db: db}
}

type AccountModel struct {
	ID              string `gorm:"column:id;primaryKey"`
	ClientID        string `gorm:"column:client_id"`
	Balance         int64  `gorm:"column:balance"`
	ReservedBalance int64  `gorm:"column:reserved_balance"`
	CreditLimit     int64  `gorm:"column:credit_limit"`
	Status          string `gorm:"column:status"`
}

func (AccountModel) TableName() string {
	return "accounts"
}

func (repo *GormAccountRepository) FindByID(ctx context.Context, id string) (*account.Account, error) {

	var accountModel AccountModel

	err := repo.db.WithContext(ctx).Where("id = ?", id).First(&accountModel).Error

	if err != nil {
		return nil, err
	}

	return &account.Account{
		ID:              accountModel.ID,
		ClientID:        accountModel.ClientID,
		Balance:         accountModel.Balance,
		ReservedBalance: accountModel.ReservedBalance,
		CreditLimit:     accountModel.CreditLimit,
		Status:          account.Status(accountModel.Status),
	}, nil
}

func (r *GormAccountRepository) Save(ctx context.Context, acc *account.Account) error {

	accountModel := AccountModel{
		ID:              acc.ID,
		ClientID:        acc.ClientID,
		Balance:         acc.Balance,
		ReservedBalance: acc.ReservedBalance,
		CreditLimit:     acc.CreditLimit,
		Status:          string(acc.Status),
	}

	// GORM requer um ponteiro (&) para operações Save/Update funcionarem perfeitamente
	return r.db.WithContext(ctx).Save(&accountModel).Error
}
