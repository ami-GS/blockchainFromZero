package utils

import (
	"strconv"

	"github.com/ami-GS/blockchainFromZero/src/transaction"
)

func ComputeChange(txInputs []transaction.TxInput, txOutputs []transaction.TxOutput, fee int) int {
	totalIn := 0
	totalOut := fee
	for _, txIn := range txInputs {
		output := txIn.GetTargetOutput()
		tmp, err := strconv.Atoi(output.Value)
		if err != nil {
			panic(err)
		}
		totalIn += tmp
	}
	for _, txOut := range txOutputs {
		tmp, err := strconv.Atoi(txOut.Value)
		if err != nil {
			panic(err)
		}
		totalOut += tmp
	}
	return totalIn - totalOut
}
