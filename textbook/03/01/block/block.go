package block

type Block struct {
	Transaction string
	PrevBlkHash string
}

func newBlock(transaction, prevBlkHash string) *Block {
	return &Block{
		Transaction: transaction,
		PrevBlkHash: prevBlkHash,
	}
}

type GenesisBlock Block

func newGenesisBlock() *GenesisBlock {
	return &GenesisBlock{
		Transaction: "deadbeefca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785fee1dead",
		PrevBlkHash: "",
	}
}
