package block

type BlockBuilder struct {
}

func (b *BlockBuilder) GenerateGenesisBlock() *GenesisBlock {
	return newGenesisBlock()
}

func (b *BlockBuilder) GenerateNewBlock(transaction, prevBlkHash string) *Block {
	return newBlock(transaction, prevBlkHash)
}
