package block

import "log"

type BlockBuilder struct {
}

func NewBlockBuilder() *BlockBuilder {
	log.Println("Initializing BlockBuilder...")
	return &BlockBuilder{}
}

func (b *BlockBuilder) GenerateGenesisBlock() *GenesisBlock {
	return newGenesisBlock()
}

func (b *BlockBuilder) GenerateNewBlock(transaction, prevBlkHash string) *Block {
	return newBlock(transaction, prevBlkHash)
}
