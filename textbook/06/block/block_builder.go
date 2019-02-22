package block

import (
	"context"
	"log"

	"github.com/ami-GS/blockchainFromZero/textbook/06/transaction"
)

type BlockBuilder struct {
}

func NewBlockBuilder() *BlockBuilder {
	log.Println("Initializing BlockBuilder...")
	return &BlockBuilder{}
}

func (b *BlockBuilder) GenerateGenesisBlock() *GenesisBlock {
	return newGenesisBlock()
}

func (b *BlockBuilder) GenerateNewBlock(transactions []transaction.Transaction, prevBlkHash []byte, ctx *context.Context) *Block {
	return newBlock(transactions, prevBlkHash, ctx)
}
