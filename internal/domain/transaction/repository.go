package transaction

import "context"

type TransactionRepository interface {
	FindByReferenceID(ctx context.Context, referenceID string) (*Transaction, error)

	Save(ctx context.Context, transaction *Transaction) error
}
