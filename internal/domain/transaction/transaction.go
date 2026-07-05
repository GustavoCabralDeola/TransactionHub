package transaction

import "time"

type Transaction struct {
	ID           string
	AccountID    string
	Operation    OperationType
	Amount       int64
	Currency     string
	ReferenceID  string
	Metadata     map[string]interface{}
	Status       Status
	ErrorMessage *string
	Timestamp    time.Time
}
