package block

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"reflect"
	"sync"

	"github.com/ami-GS/blockchainFromZero/src/block/utils"
	"github.com/ami-GS/blockchainFromZero/src/transaction"
	"github.com/pkg/errors"
)

type BlockChainManager struct {
	Chain        BlockChain // slice of pointer is not good to access items linearly
	bb           *BlockBuilder
	genesisBlock *GenesisBlock
	mu           *sync.Mutex
}

func NewBlockChainManager() *BlockChainManager {
	log.Println("Initializing BlockChain Manager...")
	bb := NewBlockBuilder()
	blk := bb.GenerateGenesisBlock()
	return &BlockChainManager{
		Chain:        []Block{Block(*blk)},
		genesisBlock: blk,
		bb:           bb,
		mu:           new(sync.Mutex),
	}
}

func (b *BlockChainManager) GenerateNewBlock(transactions []transaction.Transaction, prevBlkHash []byte, ctx *context.Context) *Block {
	return b.bb.GenerateNewBlock(transactions, prevBlkHash, ctx)
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
	return b.Chain[len(b.Chain)-1].GetHash()
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
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Chain.RemoveDupTxFromChain(transactions)
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
	// returns
	// hash []byte: previous block hash if block is updated
	// blocks : orphan blocks
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
	if err != nil {
		return nil, nil
	}
	hash, err := chain[len(chain)-1].GetHash()
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

	digest := bcutils.DoubleHashSha256(bcutils.GetBytesWithNonce(msg, nonce))
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
		prvHash, err := prvBlk.GetHash()
		if err != nil {
			panic("TODO: error")
			return false
		}
		if !b.IsValidBlock(prvHash, blk, DIFFICULTY) {
			return false
		}
		prvBlk = blk
	}
	return true
}

func (b *BlockChainManager) GetTransactions() []transaction.Transaction {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]transaction.Transaction, 0)
	for _, blk := range b.Chain {
		out = append(out, blk.Transactions...)
	}
	return out
}

func (b *BlockChainManager) HasTxOutputAsRefFromTxInput(txOut *transaction.TxOutput) bool {
	for i := 1; i < len(b.Chain); i++ {
		blk := b.Chain[i]
		for _, tx := range blk.Transactions {
			//switch tx := txI.(type) {
			//case transaction.Transaction, transaction.CoinBaseTransaction:
			for _, txIn := range tx.GetInputs() {
				refTxOut := txIn.GetTargetOutput()
				if reflect.DeepEqual(refTxOut, *txOut) {
					// not valid as TxOut
					return true
				}
			}
			//default:
			//}
		}
	}
	return false
}

// TODO: same name is used in TransactionPool as well
func (b *BlockChainManager) HasTxOut(txOut *transaction.TxOutput) bool {
	for i := 1; i < len(b.Chain); i++ {
		blk := b.Chain[i]
		for _, tx := range blk.Transactions {
			//switch tx := txI.(type) {
			//case transaction.Transaction, transaction.CoinBaseTransaction:
			for _, myTxOut := range tx.GetOutputs() {
				if reflect.DeepEqual(myTxOut, *txOut) {
					// valida as TxOut
					return true
				}
			}
			//default:
			//	}
		}
	}
	return false
}

func (b *BlockChainManager) ValidateTxOut(txOut *transaction.TxOutput) bool {
	// combination of HasTxOutputAsRefFromTxInput and HasTxOut
	log.Println("ValidateTxOut is called")
	knownTxOut := false
	for i := 1; i < len(b.Chain); i++ {
		blk := b.Chain[i]
		for _, tx := range blk.Transactions {
			//switch tx := txI.(type) {
			//case transaction.Transaction, transaction.CoinBaseTransaction:
			for _, txIn := range tx.GetInputs() {
				refTxOut := txIn.GetTargetOutputP()
				if reflect.DeepEqual(*refTxOut, *txOut) {
					log.Printf("TxOut is already used : \n%s\n%s\n", *txOut, *refTxOut)
					// ->
					log.Println("NOTICE: =============================================> Temporally skipping for debug")
					//return false
					// <-
				}
			}
			if !knownTxOut {
				for _, myTxOut := range tx.GetOutputs() {
					if reflect.DeepEqual(myTxOut, *txOut) {
						knownTxOut = true
					}
				}
			}
			//default:
			//}
		}
	}
	if !knownTxOut {
		log.Printf("Not observed TxOut : %v\n", *txOut)
		// ->
		log.Println("NOTICE =============================================> Temporally return true for debug")
		return true
		// <-
	}

	return knownTxOut
}
