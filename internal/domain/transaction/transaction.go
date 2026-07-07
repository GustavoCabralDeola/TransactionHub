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

// Faz com que toda transação venha obrigatóriamente no status "pending"
func NewTransaction(id, accountID string, operationType OperationType, amount int64,
	currency, referenceID string, metadata map[string]interface{}) *Transaction {
	return &Transaction{
		ID:          id,
		AccountID:   accountID,
		Operation:   operationType,
		Amount:      amount,
		Currency:    currency,
		ReferenceID: referenceID,
		Metadata:    metadata,
		Status:      Pending,
		Timestamp:   time.Now(),
	}

}

func (t *Transaction) MarkAsSucess() {
	t.Status = Sucess
	t.Timestamp = time.Now()
}

func (t *Transaction) MarkAsFailed(reason string) {
	t.Status = Failed
	t.ErrorMessage = &reason
	t.Timestamp = time.Now()

}
