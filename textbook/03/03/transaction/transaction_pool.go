package transaction

import (
	"encoding/json"
	"log"
	"sync"
)

type TransactionPool struct {
	transactions []string
	mu           *sync.Mutex
}

func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		transactions: make([]string, 0),
		mu:           new(sync.Mutex),
	}
}

func (t *TransactionPool) Append(trans string) {
	t.mu.Lock()
	t.transactions = append(t.transactions, trans)
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
	t.transactions = make([]string, 0)
	t.mu.Unlock()
}

func (t *TransactionPool) Get() (int, []byte) {
	if len(t.transactions) == 0 {
		log.Println("TransactionPool.Get() is called, but the pool is empty")
		return 0, nil
	}

	t.mu.Lock()
	outLen := len(t.transactions)
	out, err := json.Marshal(t.transactions)
	t.mu.Unlock()
	if err != nil {
		panic(err)
	}
	return outLen, out
}
