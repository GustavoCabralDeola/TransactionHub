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
	ID              string `gorm: "primary_key; column:id"`
	Balance         int64  `gorm: "column:balance"`
	ReservedBalance int64  `gorm: "column:reserved_balance"`
	CreditLimit     int64  `gorm: "column:credit_limit"`
	Status          string `gorm: "column:status"`
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
		Balance:         accountModel.Balance,
		ReservedBalance: accountModel.ReservedBalance,
		CreditLimit:     accountModel.CreditLimit,
		Status:          account.Status(accountModel.Status),
	}, nil
}

func (r *GormAccountRepository) Save(ctx context.Context, account *account.Account) error {

	accountModel := AccountModel{

		ID:              account.ID,
		Balance:         account.Balance,
		ReservedBalance: account.ReservedBalance,
		CreditLimit:     account.CreditLimit,
		Status:          string(account.Status),
	}

	return r.db.WithContext(ctx).Save(accountModel).Error
}
