package block

import (
	"crypto/sha256"
	"encoding/json"
	"log"
	"sync"

	"github.com/pkg/errors"
)

type BlockChainManager struct {
	Chain        []Block // slice of pointer is not good to access items linearly
	genesisBlock *GenesisBlock
	mu           *sync.Mutex
}

func NewBlockChainManager(blk *GenesisBlock) *BlockChainManager {
	return &BlockChainManager{
		Chain:        []Block{Block(*blk)},
		genesisBlock: blk,
		mu:           new(sync.Mutex),
	}
}

func (b *BlockChainManager) setGenesisBlock(blk *GenesisBlock) {
	b.genesisBlock = blk
	b.Chain = append(b.Chain, Block(*blk))
}

func (b *BlockChainManager) AppendNewBlock(blk *Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Chain = append(b.Chain, *blk)
}

func (b *BlockChainManager) IsValidChain(chain []Block) bool {
	if len(chain) == 0 {
		return false
	}
	prvBlk := chain[0]
	for i := 1; i < len(chain); i++ {
		blk := chain[i]
		prvHash, err := b.GetHash(&prvBlk)
		if err != nil {
			log.Println(err)
			return false
		}
		if blk.PrevBlkHash != string(prvHash) {
			return false
		}
		prvBlk = blk
	}
	return true
}

func (b *BlockChainManager) doubleHashSha256(data []byte) []byte {
	tmp := sha256.Sum256(data)
	tmp = sha256.Sum256(tmp[:])
	return tmp[:]
}

func (b *BlockChainManager) GetHash(blk *Block) ([]byte, error) {
	jsonBlk, err := json.Marshal(blk)
	if err != nil {
		return nil, errors.Wrap(err, "failed to jasonize block")
	}
	return b.doubleHashSha256(jsonBlk), nil
}
