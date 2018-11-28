package main

import (
	"encoding/json"
	"fmt"

	"github.com/ami-GS/blockchainFromZero/textbook/03/01/block"
)

func main() {
	bb := block.BlockBuilder{}
	genesisBlock := bb.GenerateGenesisBlock()
	bm := block.NewBlockChainManager(genesisBlock)

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
	newBlock2 := bb.GenerateNewBlock(string(jsonTransaction2), string(newBlkHash))
	bm.AppendNewBlock(newBlock2)

	fmt.Println("IsValidChain:", bm.IsValidChain(bm.Chain))
}
