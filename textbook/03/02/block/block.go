package block

import "time"

type Block struct {
	Timestamp    time.Time
	Transactions string
	PrevBlkHash  string
}

func newBlock(transaction, prevBlkHash string) *Block {
	return &Block{
		Timestamp:    time.Now(),
		Transactions: transaction,
		PrevBlkHash:  prevBlkHash,
	}
}

type GenesisBlock Block

func newGenesisBlock() *GenesisBlock {
	return &GenesisBlock{
		Timestamp:    time.Now(),
		Transactions: "deadbeefca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785fee1dead",
		PrevBlkHash:  "",
	}
}
