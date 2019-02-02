package utxo

import (
	"log"
	"reflect"
	"strconv"

	"github.com/ami-GS/blockchainFromZero/textbook/04/transaction"
)

type UTXOManager struct {
	Address    string // Pubkey?
	Balance    int
	InputsToMe []transaction.TxInput
}

func NewUTXOManager(address string) *UTXOManager {
	return &UTXOManager{
		Address:    address,
		Balance:    0,
		InputsToMe: make([]transaction.TxInput, 0),
	}
}

func (u *UTXOManager) ExtractUTXO(txs []transaction.Transaction) {
	// TODO: need lock?
	// TODO: could be optimized
	log.Println("ExtractUTXO is called")
	outTxsToMe := make([]transaction.Transaction, 0)
	for _, tx := range txs {
		for _, txOut := range tx.Outputs {
			if txOut.Recipient == u.Address {
				outTxsToMe = append(outTxsToMe, tx)
			}
		}
	}

	if len(outTxsToMe) == 0 {
		log.Println("No Transactions for UTXO")
		return
	}
	results := make([]transaction.Transaction, 0)
	for _, tx := range txs {
		for _, txIn := range tx.Inputs {
			outputInInputTx := txIn.GetTargetOutput()
			if u.Address == outputInInputTx.Recipient {
				for _, outTxToMe := range outTxsToMe {
					if !reflect.DeepEqual(txIn.Tx, outTxToMe) {
						results = append(results, outTxToMe)
					}
				}
			}
		}
	}
	// TODO: not good
	if len(results) == 0 {
		u.setUTXOTxs(outTxsToMe)
	} else {
		u.setUTXOTxs(results)
	}
}

func (u *UTXOManager) setUTXOTxs(txs []transaction.Transaction) {
	// TODO: this capacity might not be correct
	inputsToMe := make([]transaction.TxInput, 0, len(txs))
	for _, tx := range txs {
		for idx, output := range tx.Outputs {
			if output.Recipient == u.Address {
				inputsToMe = append(inputsToMe, transaction.TxInput{tx, idx})
			}
		}
	}
	u.InputsToMe = inputsToMe

	u.ComputeBalance()
}

func (u *UTXOManager) ComputeBalance() {
	balance := 0
	for _, txInput := range u.InputsToMe {
		//for _, txOut := range txInput.Tx
		txOut := txInput.GetTargetOutput()
		tmp, err := strconv.Atoi(txOut.Value)
		if err != nil {
			panic(err)
		}
		balance += tmp
	}
	u.Balance = balance
}

func (u *UTXOManager) RemoveInputsFromBeginning(toIdx int) {
	// TODO: need index overrun check
	u.InputsToMe = u.InputsToMe[toIdx:]
}
