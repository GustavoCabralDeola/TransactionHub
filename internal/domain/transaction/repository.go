package transaction

import "context"

type Repository interface {
	FindByID(ctx context.Context, id string) (*Transaction, error)
	FindByReferenceID(ctx context.Context, referenceID string) (*Transaction, error)
	Save(ctx context.Context, transaction *Transaction) error
}
