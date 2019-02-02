package transaction

import (
	"encoding/json"
	"strconv"
	"time"
)

type TxInput struct {
	Tx          Transaction
	OutputIndex int
}

func (t *TxInput) GetTargetOutput() TxOutput {
	return t.Tx.Outputs[t.OutputIndex]
}

type TxOutput struct {
	Recipient string // or []byte?
	Value     string
}

type Transaction struct {
	Inputs    []TxInput
	Outputs   []TxOutput
	TimeStamp time.Time
	Sign      []byte
}

func New(inputs []TxInput, outputs []TxOutput) *Transaction {
	t := &Transaction{
		Inputs:    inputs,
		Outputs:   outputs,
		TimeStamp: time.Now(),
	}
	// if !t.HasEnoughInputs() {
	// 	return nil, erorrs.Wrap(nil, "output values exceed total balance")
	// }
	return t
}

func (t *Transaction) HasEnoughInputs(fee int) bool {
	totalIn := 0
	for _, inTx := range t.Inputs {
		val, err := strconv.Atoi(inTx.GetTargetOutput().Value)
		if err != nil {
			panic(err)
		}
		totalIn += val
	}

	totalOut := fee
	for _, oTx := range t.Outputs {
		val, err := strconv.Atoi(oTx.Value)
		if err != nil {
			panic(err)
		}
		totalOut += val
	}

	return totalIn >= totalOut
}

type CoinBaseTransaction Transaction

func NewCoinBaseTransaction(recipient string, value int) *CoinBaseTransaction {
	val := strconv.Itoa(value)

	txout := TxOutput{
		Recipient: recipient,
		Value:     val,
	}

	return (*CoinBaseTransaction)(New(nil, []TxOutput{txout}))
}

func (t *Transaction) GetJson() ([]byte, error) {
	return json.Marshal(*t)
}
