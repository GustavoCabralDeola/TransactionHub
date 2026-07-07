package account

import "errors"

type Status string

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrAccountBlocked      = errors.New("account is blocked or inactive")
)

const (
	Active   Status = "active"
	Inactive Status = "inactive"
	Blocked  Status = "blocked"
)
