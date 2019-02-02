package transaction

import (
	"log"
	"reflect"
	"sync"
)

type TransactionPool struct {
	transactions []Transaction
	mu           *sync.Mutex
}

func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		transactions: make([]Transaction, 0),
		mu:           new(sync.Mutex),
	}
}

func (t *TransactionPool) Append(trans ...Transaction) {
	t.mu.Lock()
	t.transactions = append(t.transactions, trans...)
	t.mu.Unlock()
}

func (t *TransactionPool) PopFront(index int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if 0 <= index && index <= len(t.transactions) {
		t.transactions = t.transactions[index:]
	}
}

func (t *TransactionPool) Flush() {
	t.mu.Lock()
	t.transactions = make([]Transaction, 0)
	t.mu.Unlock()
}

func (t *TransactionPool) Get() ([]Transaction, int) {
	if len(t.transactions) == 0 {
		log.Println("TransactionPool.Get() is called, but the pool is empty")
		return nil, 0
	}

	return t.transactions, len(t.transactions)
}

func (t *TransactionPool) GetCopy() ([]Transaction, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	num := len(t.transactions)
	if num == 0 {
		log.Println("TransactionPool.GetCopy() is called, but the pool is empty")
		return nil, 0
	}
	out := make([]Transaction, num)
	copy(out, t.transactions)
	return out, num
}

func (t *TransactionPool) Has(inTx Transaction) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, tx := range t.transactions {
		if reflect.DeepEqual(inTx, tx) {
			return true
		}
	}
	return false
}

func (t *TransactionPool) TrimTransactions(transactions []Transaction) {
	t.mu.Lock()
	defer t.mu.Unlock()
	myTxs := t.transactions
	t.transactions = make([]Transaction, 0, len(t.transactions))
	for _, myTx := range myTxs {
		for _, commingTx := range transactions {
			if reflect.DeepEqual(myTx, commingTx) {
				goto INCLUDE
			}
		}
		t.transactions = append(t.transactions, myTx)
	INCLUDE:
	}
}

func (t *TransactionPool) OverwriteTransactions(transactions []Transaction) {
	t.mu.Lock()
	defer t.mu.Unlock()
	log.Println("OverwriteTransactions is called")
	log.Println("\tcurrent transaction pool is", t.transactions)
	log.Println("\ttransaction pool will be overwritten by", transactions)
	if transactions == nil {
		t.transactions = make([]Transaction, 0)
	} else {
		t.transactions = transactions
	}
}
