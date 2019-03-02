package block

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/ami-GS/blockchainFromZero/src/block/utils"
	"github.com/ami-GS/blockchainFromZero/src/transaction"
	"github.com/pkg/errors"
)

type Block struct {
	Timestamp    time.Time
	Transactions []transaction.Transaction
	PrevBlkHash  []byte
	Nonce        uint64
}

func (b Block) String() string {
	TxsString := ""
	for i, tx := range b.Transactions {
		TxsString += fmt.Sprintf("%d : %s", i, tx)
	}

	return fmt.Sprintf(`
	Timestamp: %v
	Txs: %s
	Previous hash: %s
	Nonce: %d
`, b.Timestamp, TxsString, string(b.PrevBlkHash), b.Nonce)
}

// TODO: to be dynamic
const DIFFICULTY = 2

func (b *Block) Equal(right *Block) bool {
	if len(b.Transactions) != len(right.Transactions) {
		return false
	}
	for i, tx := range b.Transactions {
		// This checks order as well
		if reflect.DeepEqual(tx, right.Transactions[i]) {
			return false
		}
	}
	return b.Timestamp == right.Timestamp &&
		bytes.Equal(b.PrevBlkHash, right.PrevBlkHash) &&
		b.Nonce == right.Nonce &&
		len(b.Transactions) == len(right.Transactions)
}

func (b *Block) GetTotalFee() int {
	totalFee := 0
	for _, tx := range b.Transactions {
		totalFee += tx.GetFee()
	}
	return totalFee
}

func (b *Block) GetHash() ([]byte, error) {
	jsonBlk, err := json.Marshal(b)
	if err != nil {
		return nil, errors.Wrap(err, "failed to jasonize block")
	}
	return bcutils.DoubleHashSha256(jsonBlk), nil
}

func newBlock(transactions []transaction.Transaction, prevBlkHash []byte, ctx *context.Context) *Block {
	// TODO: copy is faster
	txs := append([]transaction.Transaction{}, transactions...)
	blk := &Block{
		Timestamp:    time.Now(),
		Transactions: txs,
		PrevBlkHash:  prevBlkHash,
		Nonce:        0,
	}

	json, err := json.Marshal(blk)
	if err != nil {
		panic(err)
	}
	if ctx != nil {
		blk.Nonce = bcutils.ComputeNonceForPowWithCancel(json, DIFFICULTY, *ctx)
		if blk.Nonce == 0 {
			return nil
		}
	}
	return blk
}

type GenesisBlock Block

func newGenesisBlock() *GenesisBlock {
	genesisTx := transaction.New(
		[]transaction.TxInput{},
		[]transaction.TxOutput{
			transaction.TxOutput{[]byte("deadbeefca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785fee1dead"), ""},
		})
	genesisTx.TimeStamp = time.Time{} // means 0
	return &GenesisBlock{
		Timestamp:    time.Time{}, // means 0
		Transactions: []transaction.Transaction{*genesisTx},
		PrevBlkHash:  nil,
	}
}
