package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/03/02/block"
	"github.com/ami-GS/blockchainFromZero/textbook/03/02/transaction"
)

func generateBlockWithTP(tp *transaction.TransactionPool, bb *block.BlockBuilder, bm *block.BlockChainManager, prvBlkHash []byte) {
	txsJson := tp.Get()
	if txsJson == nil {
		fmt.Println("Transaction Pool is empty")
	}
	newBlk := bb.GenerateNewBlock(string(txsJson), string(prvBlkHash))
	bm.AppendNewBlock(newBlk)
	prvBlkHash, err := bm.GetHash(newBlk)
	if err != nil {
		panic(err)
	}
	tp.Flush()
	fmt.Println("Current Blockchain is ...", bm.Chain)
	fmt.Println("Current prvBlkHash is ...", prvBlkHash)
}

func main() {
	bb := block.BlockBuilder{}
	genesisBlock := bb.GenerateGenesisBlock()
	bm := block.NewBlockChainManager(genesisBlock)
	tp := transaction.NewTransactionPool()

	prvBlkHash, err := bm.GetHash((*block.Block)(genesisBlock))
	if err != nil {
		panic(err)
	}
	fmt.Println("Genesis Block Hash:", string(prvBlkHash))
	transaction := map[string]string{
		"Sender":    "test1",
		"Recipient": "test2",
		"Value":     "333",
	}

	jsonTransaction, err := json.Marshal(transaction)
	if err != nil {
		panic(err)
	}

	tp.Append(string(jsonTransaction))

	newBlock := bb.GenerateNewBlock(string(jsonTransaction), string(prvBlkHash))
	bm.AppendNewBlock(newBlock)
	newBlkHash, err := bm.GetHash(newBlock)
	if err != nil {
		panic(err)
	}

	fmt.Println("1st Block Hash:", string(newBlkHash))

	transaction2 := map[string]string{
		"Sender":    "test1",
		"Recipient": "test2",
		"Value":     "222",
	}
	jsonTransaction2, err := json.Marshal(transaction2)
	if err != nil {
		panic(err)
	}
	tp.Append(string(jsonTransaction2))

	ticker := time.NewTicker(10 * time.Second)
	timer := time.NewTimer(15 * time.Second)
OUT:
	for {
		select {
		case <-ticker.C:
			go generateBlockWithTP(tp, &bb, bm, prvBlkHash)
		case <-timer.C:
			break OUT
		}
	}
	transaction3 := map[string]string{
		"Sender":    "test1",
		"Recipient": "test2",
		"Value":     "111",
	}
	jsonTransaction3, err := json.Marshal(transaction3)
	if err != nil {
		panic(err)
	}
	tp.Append(string(jsonTransaction3))
	fmt.Println(*tp)
}
