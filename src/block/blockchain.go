package block

import (
	"fmt"
	"reflect"

	"github.com/ami-GS/blockchainFromZero/src/transaction"
)

type BlockChain []Block

func (b BlockChain) String() string {
	out := "\n======================= Blockchain begin ===========================\n"
	out += fmt.Sprintf("[ BLOCK 1 \n%s\n]", b[0].String())
	for i := 1; i < len(b); i++ {
		out += fmt.Sprintf(" -> [ BLOCK %d\n%s\n]", i+1, b[i].String())
	}
	out += "\n======================= Blockchain end ===========================\n"
	return out
}

func (chain *BlockChain) RemoveDupTxFromChain(transactions []transaction.Transaction) []transaction.Transaction {
	if len(transactions) == 0 {
		return nil
	}

	// TODO: if the transactions comes from orphan block know the block idx, this traversal could be shortcutted
	out := make([]transaction.Transaction, 0, len(transactions))
	for _, t1 := range transactions {
		for i := 1; i < len(*chain); i++ {
			txs := (*chain)[i].Transactions
			for _, t2 := range txs {
				if reflect.DeepEqual(t1, t2) {
					goto INCLUDE
				}
			}
		}
		out = append(out, t1)
	INCLUDE:
	}
	return out
}
