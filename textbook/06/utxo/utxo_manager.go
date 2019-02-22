package utxo

import (
	"bytes"
	"log"
	"strconv"

	"github.com/ami-GS/blockchainFromZero/textbook/06/transaction"
)

type UTXOManager struct {
	Address    []byte // Pubkey?
	Balance    int
	InputsToMe []transaction.TxInput
}

func NewUTXOManager(address []byte) *UTXOManager {
	return &UTXOManager{
		Address:    address,
		Balance:    0,
		InputsToMe: make([]transaction.TxInput, 0),
	}
}

func (u *UTXOManager) ExtractUTXO(txs []transaction.Transaction) []transaction.Transaction {
	// TODO: can be optimize using hashMap
	log.Println("ExtractUTXO is called")
	retTxOuts := make([]transaction.Transaction, 0)
	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		for _, txOut := range tx.GetOutputs() {
			// check if TxOutput is to me
			if bytes.Equal(txOut.Recipient, u.Address) {
				txIns := tx.GetInputs()
				// check if it is coinbase
				if len(txIns) > 0 {
					for _, txIn := range txIns {
						usedOut := txIn.GetTargetOutput()
						usedTx := txIn.Tx
						// check if used TxOut is to me (means that send from me to me (change) )
						removeIdx := make([]int, 0)
						if bytes.Equal(usedOut.Recipient, u.Address) {
							// this is change to sender
							// remove stored Tx
							for txtxIdx, txtx := range retTxOuts {
								if usedTx.EqualWithoutSign(&txtx) {
									removeIdx = append(removeIdx, txtxIdx)
								}
							}
						} else {
							outLen := len(retTxOuts)
							// Avoid to append same tx, this happens when there are several TxInput
							if outLen == 0 || (outLen > 0 && !retTxOuts[outLen-1].EqualWithoutSign(&tx)) {
								retTxOuts = append(retTxOuts, tx)
							}
						}
						if len(removeIdx) != 0 {
							for i := len(removeIdx) - 1; i >= 0; i-- {
								retTxOuts = append(retTxOuts[:i], retTxOuts[i+1:]...)
							}
							// append tx which include change to me
							outLen := len(retTxOuts)
							if outLen == 0 || (outLen > 0 && !retTxOuts[outLen-1].EqualWithoutSign(&tx)) {
								retTxOuts = append(retTxOuts, tx)
							}
						}
					}
				} else { // for coinbase
					retTxOuts = append(retTxOuts, tx)
				}
			}
		}
		//CONTINUE:
	}
	return retTxOuts
}

func (u *UTXOManager) SetUTXOTxs(txs []transaction.Transaction) {
	// TODO: this capacity might not be correct
	inputsToMe := make([]transaction.TxInput, 0, len(txs))
	for _, tx := range txs {
		for idx, output := range tx.GetOutputs() {
			if bytes.Equal(output.Recipient, u.Address) {
				inputsToMe = append(inputsToMe, transaction.TxInput{tx, idx})
			}
		}
	}
	u.InputsToMe = inputsToMe
}

func (u *UTXOManager) ComputeBalance() int {
	log.Println("ComputeBalance is called")
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
	log.Println("Current balance is :", balance)
	return balance
}

func (u *UTXOManager) RemoveInputsFromBeginning(toIdx int) {
	// TODO: need index overrun check
	u.InputsToMe = u.InputsToMe[toIdx:]
}
