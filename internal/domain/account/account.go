package account

import (
	"errors"
	"time"
)

type Account struct {
	ID              string
	ClientID        string
	Balance         int64
	ReservedBalance int64
	CreditLimit     int64
	Status          Status
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Calcula o saldo disponível na conta em tempo real
func (a *Account) AvailableBalance() int64 {

	return a.Balance - a.ReservedBalance
}

// Adiciona valor ao saldo da conta
func (a *Account) Credit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.Status != Active {
		return ErrAccountBlocked
	}

	a.Balance += amount
	a.UpdatedAt = time.Now()

	return nil
}

// Remove valor do saldo da conta
func (a *Account) Debit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.Status != Active {
		return ErrAccountBlocked
	}

	if amount > (a.AvailableBalance() + a.CreditLimit) {
		return ErrInsufficientBalance
	}

	a.Balance -= amount
	a.UpdatedAt = time.Now()

	return nil
}

// Move valor do saldo disponível para o saldo reservado
func (a *Account) Reserve(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if a.Status != Active {
		return ErrAccountBlocked
	}

	if amount > a.AvailableBalance() {
		return ErrInsufficientBalance
	}

	a.ReservedBalance += amount
	a.UpdatedAt = time.Now()
	return nil

}

// Confirma uma reserva, removendo do saldo reservado
func (a *Account) Capture(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if amount > a.ReservedBalance {
		return errors.New("insufficient reserved balance")
	}

	if a.Status != Active {
		return ErrAccountBlocked
	}

	a.ReservedBalance -= amount
	a.Balance -= amount
	a.UpdatedAt = time.Now()

	return nil
}
