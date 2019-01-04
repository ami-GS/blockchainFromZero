package transaction

import (
	"encoding/json"
	"strconv"
)

type Transaction struct {
	Sender    string
	Recipient string
	Value     string
}

func New(sender, recipient string, value interface{}) *Transaction {
	v := ""
	switch actual := value.(type) {
	case string:
		v = actual
	case int:
		v = strconv.Itoa(actual)
	default:
		panic("")
	}

	return &Transaction{
		Sender:    sender,
		Recipient: recipient,
		Value:     v,
	}
}

func (t *Transaction) GetJson() ([]byte, error) {
	return json.Marshal(*t)
}
