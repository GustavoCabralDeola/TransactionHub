package commands

import "sync"

var (
	accountLocks sync.Map
)

// LockAccount cria um bloqueio (Mutex) para uma conta específica
func LockAccount(accountID string) func() {
	val, _ := accountLocks.LoadOrStore(accountID, &sync.Mutex{})
	mu := val.(*sync.Mutex)

	mu.Lock()

	return func() {
		mu.Unlock()
	}
}

// LockTransferAccounts trava duas contas simultaneamente.
func LockTransferAccounts(sourceID, destinationID string) func() {
	if sourceID == destinationID {
		return LockAccount(sourceID)
	}

	first, second := sourceID, destinationID
	if first > second {
		first, second = second, first
	}

	unlockFirst := LockAccount(first)
	unlockSecond := LockAccount(second)

	return func() {
		unlockSecond()
		unlockFirst()
	}
}
