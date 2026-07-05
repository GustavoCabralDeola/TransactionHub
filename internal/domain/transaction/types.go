package transaction

type OperationType string

const (
	Credit   OperationType = "credit"
	Debit    OperationType = "debit"
	Reserve  OperationType = "reserve"
	Capture  OperationType = "capture"
	Reversal OperationType = "reversal"
	Transfer OperationType = "transfer"
)

type Status string

const (
	Pending Status = "pending"
	Sucess  Status = "success"
	Failed  Status = "failed"
)
