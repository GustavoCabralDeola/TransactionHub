package account

import "context"

type Repository interface {
	FindByID(ctx context.Context, ID string) (*Account, error)

	Save(ctx context.Context, a *Account) error
}
