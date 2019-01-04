package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"log"
	"reflect"
	"strconv"
	"sync"

	"github.com/ami-GS/blockchainFromZero/textbook/03/04/transaction"
	"github.com/pkg/errors"
)

type BlockChainManager struct {
	Chain        BlockChain // slice of pointer is not good to access items linearly
	genesisBlock *GenesisBlock
	mu           *sync.Mutex
}

func NewBlockChainManager(blk *GenesisBlock) *BlockChainManager {
	log.Println("Initializing BlockChain Manager...")
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

func (b *BlockChainManager) SetChain(chain BlockChain) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.IsValidChain(chain) {
		log.Println("Invalid chain was passed for SetChain(chain)")
		return nil, errors.Wrap(nil, "Invalid chain")
	}
	b.Chain = chain
	return b.GetHash(&b.Chain[len(b.Chain)-1])
}

func (b *BlockChainManager) GetTransactionsFromOrphanBlocks(orphanBlocks []Block) []transaction.Transaction {
	notProcessedTxs := make([]transaction.Transaction, 0)
	for _, orphanBlock := range orphanBlocks {
		uniqueTxs := b.RemoveDuplicateTransactions(orphanBlock.Transactions)
		notProcessedTxs = append(notProcessedTxs, uniqueTxs...)
	}
	return notProcessedTxs
}

func (b *BlockChainManager) RemoveDuplicateTransactions(transactions []transaction.Transaction) []transaction.Transaction {
	if len(transactions) == 0 {
		return nil
	}

	// TODO: if the transactions comes from orphan block know the block idx, this traversal could be shortcutted
	out := make([]transaction.Transaction, 0, len(transactions))
	for _, t1 := range transactions {
		for i := 1; i < len(b.Chain); i++ {
			txs := b.Chain[i].Transactions
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

func (b *BlockChainManager) RenewChainBy(chain BlockChain) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.IsValidChain(chain) {
		log.Println("A chain for renewing is invalid")
		// This can be ignored
		return errors.Wrap(nil, "A chain for renewing is invalid")
	}
	b.Chain = chain
	return nil
}

// TODO: name is not appropriate
func (b *BlockChainManager) ResolveBranch(chain BlockChain) ([]byte, []Block) {
	myLen := len(b.Chain)
	newLen := len(chain)

	myChain := append([]Block{}, b.Chain...) // avoid update during this resolution
	if newLen <= myLen {
		log.Println("Shorter chain incomming, mine is correct")
		return nil, nil
	}
	returnBlocks := make([]Block, 0, len(myChain))
	for _, myBlk := range myChain {
		for _, newBlk := range chain {
			if myBlk.Equal(&newBlk) {
				goto INCLUDE
			}
		}
		returnBlocks = append(returnBlocks, myBlk)
	INCLUDE:
	}
	err := b.RenewChainBy(chain)
	if err == nil {
		return nil, nil
	}
	hash, err := b.GetHash(&chain[len(chain)-1])
	if err != nil {
		return nil, nil
	}
	return hash, returnBlocks
}

func (b *BlockChainManager) IsValidBlock(prvHash []byte, blk Block, difficulty int) bool {
	// blk is copy, avoiding update
	nonce := blk.Nonce
	blk.Nonce = 0
	msg, err := json.Marshal(blk)
	if err != nil {
		panic(err)
		return false
	}

	if !bytes.Equal(blk.PrevBlkHash, prvHash) {
		log.Println("Invalid block: bad previous block hash")
		return false
	}

	digest := DoubleHashSha256(GetBytesWithNonce(msg, nonce))
	if bytes.Equal(digest[len(digest)-difficulty:], make([]byte, difficulty)) {
		log.Printf("Valid block: %v\n", blk)
		return true
	}
	log.Printf("Invalid block: bad nonce")
	return false

}

func (b *BlockChainManager) IsValidChain(chain BlockChain) bool {
	if len(chain) == 0 {
		return false
	}
	prvBlk := chain[0]
	for i := 1; i < len(chain); i++ {
		blk := chain[i]
		prvHash, err := b.GetHash(&prvBlk)
		if err != nil {
			panic("TODO: error")
			return false
		}
		if !b.IsValidBlock(prvHash, blk, 3) {
			return false
		}
		prvBlk = blk
	}
	return true
}

// TODO: goto utility
func DoubleHashSha256(data []byte) []byte {
	tmp := sha256.Sum256(data)
	tmp = sha256.Sum256(tmp[:])
	return tmp[:]
}

// TODO: goto utility
func GetBytesWithNonce(msg []byte, nonce uint64) []byte {
	// TODO: can be optimized
	return append(msg, []byte(strconv.FormatUint(nonce, 10))...)
}

func (b *BlockChainManager) GetHash(blk *Block) ([]byte, error) {
	jsonBlk, err := json.Marshal(blk)
	if err != nil {
		return nil, errors.Wrap(err, "failed to jasonize block")
	}
	return DoubleHashSha256(jsonBlk), nil
}
