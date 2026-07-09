package dto

import "time"

type TransactionResponse struct {
	TransactionID    string    `json:"transaction_id"`
	AccountID        string    `json:"account_id"`
	Operation        string    `json:"operation"`
	Amount           int64     `json:"amount"`
	Currency         string    `json:"currency"`
	ReferenceID      string    `json:"reference_id"`
	Status           string    `json:"status"`
	Balance          int64     `json:"balance"`
	ReservedBalance  int64     `json:"reserved_balance"`
	AvailableBalance int64     `json:"available_balance"`
	ErrorMessage     *string   `json:"error_message"`
	Timestamp        time.Time `json:"timestamp"`
}

type AccountResponse struct {
	ID               string `json:"id"`
	ClientID         string `json:"client_id"`
	Balance          int64  `json:"balance"`
	ReservedBalance  int64  `json:"reserved_balance"`
	AvailableBalance int64  `json:"available_balance"`
	CreditLimit      int64  `json:"credit_limit"`
	Status           string `json:"status"`
}
