package core

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/07/block"
	"github.com/ami-GS/blockchainFromZero/textbook/07/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/07/p2p/message"
	"github.com/ami-GS/blockchainFromZero/textbook/07/transaction"
	"github.com/pkg/errors"
)

type ServerCore struct {
	*Core
	protocolMessageStore map[string]struct{} // store payload, duplication check
	tp                   *transaction.TransactionPool
	genBlockLoopCancel   context.CancelFunc
	genBlockInterval     time.Duration
}

func NewServerCore(port uint16, bootStrapNode *p2p.Node) *ServerCore {
	log.Println("Initialize server ...")
	s := &ServerCore{
		protocolMessageStore: make(map[string]struct{}),
		tp:                   transaction.NewTransactionPool(),
		// TODO: in block module?
		genBlockInterval: 5 * time.Second,
	}
	s.Core = newCore(port, bootStrapNode, s.handleMessage, true)
	return s
}

func (s *ServerCore) Start() (context.Context, context.CancelFunc) {
	s.State = STATE_STANBY
	s.Core.Start()
	go s.generateBlockLoop()
	return s.coreContext, s.coreCancel
}

// TODO: in block module?
func (s *ServerCore) generateBlockLoop() {
	thisCtx, cancel := context.WithCancel(s.coreContext)
	defer cancel()
	s.genBlockLoopCancel = cancel

	log.Printf("Generate block with transactions")
	logPrintBlockInfo := func() {
		log.Println("Current Blockchain is ...", s.bm.Chain)
		log.Println("Current prvBlkHash is ...", string(s.prvBlkHash))
	}
	logPrintBlockInfo()
	genblock := func() *block.Block {

		//txs, num := s.tp.Get()
		txs, _ := s.tp.GetCopy()
		prvHash := make([]byte, len(s.prvBlkHash))
		// TODO: not perfectly safe
		copy(prvHash, s.prvBlkHash)
		if txs == nil {
			return nil
		}
		// TODO: suspicious ->
		trimmedTxs := s.bm.RemoveDuplicateTransactions(txs)
		numTxs := len(trimmedTxs)
		s.tp.OverwriteTransactions(trimmedTxs)
		if len(trimmedTxs) == 0 {
			return nil
		}
		// <-

		fee := s.tp.GetTotalFee()
		// TODO: not here
		fee += 2
		coinbaseTx := transaction.NewCoinBaseTransaction(s.GetPublicKeyBytes(), fee)
		// WARN: this is stopped if valid NEW_BLOCK comes
		trimmedTxs = append(trimmedTxs, *coinbaseTx)
		newBlk := s.bb.GenerateNewBlock(trimmedTxs, prvHash, &thisCtx)
		if newBlk == nil {
			// nil means process canceled
			return nil
		}
		log.Println("*************************************** Block Nonce Generated *******************************************")

		if !bytes.Equal(prvHash, s.prvBlkHash) {
			log.Println("Mining failed : Someone might win PoW earlier than you")
			return nil
		}

		s.bm.AppendNewBlock(newBlk)
		prvBlkHash, err := s.bm.GetHash(newBlk)
		if err != nil {
			panic(err)
		}
		s.prvBlkHash = prvBlkHash
		s.tp.PopFront(numTxs)
		logPrintBlockInfo()
		return newBlk
	}
	genMsgAndSend := func(blk *block.Block) {
		if blk == nil {
			return
		}
		err := s.Broadcast(blk, message.NEW_BLOCK)
		if err != nil {
			log.Println("failed to send NEW_BLOCK message :", blk)
		}
	}

	ticker := time.NewTicker(s.genBlockInterval)
	processing := false
	for {
		select {
		case <-thisCtx.Done():
			return
		case <-ticker.C:
			// TODO: ticker should not appropriate, need to consider genblock speed
			if !processing {
				go func() {
					processing = true
					genMsgAndSend(genblock())
					processing = false
				}()
			}
		}
	}
}

func (s *ServerCore) JoinNetwork() error {
	if s.BootstrapCoreNode == nil {
		log.Println("This server is running as genesis core node ...")
		return nil
	}
	s.State = STATE_CONNECT_TO_NETWORK
	return s.cm.JoinNetwork(s.BootstrapCoreNode, nil)
}

func (s *ServerCore) Shutdown() {
	s.State = STATE_SHUTTING_DOWN
	log.Println("Shutdown server ...")
	s.Core.Shutdown()
}

func (s *ServerCore) API(request message.RequestType, msg *message.Message) (message.APIResponseType, error) {
	switch request {
	case message.GET_API_ORIGIN:
		return message.SERVER_CORE_API, nil
	case message.SEND_TO_ALL_PEER:
		s.cm.SendMsgToAllPeer(msg.Payload, message.ENHANCED)
		return message.API_OK, nil
	case message.SEND_TO_ALL_EDGE:
		s.cm.SendMsgToAllEdge(msg.Payload, message.ENHANCED)
		return message.API_OK, nil
	case message.SEND_TO:
		enhancedMsg := make(map[string]string)
		json.Unmarshal(msg.Payload, &enhancedMsg)
		to := enhancedMsg["To"]
		toIP, toPort, err := s.cm.GetEdgeByPubkey(to)
		if err != nil {
			// could not find, broadcast from message_handler.go
			return message.API_ERROR, err
		}
		s.cm.SendMsgTo(msg.Payload, message.ENHANCED, &p2p.Node{toIP, uint16(toPort), nil})
		log.Println("Sending direct message to ", to)
		return message.API_OK, nil
	default:
		return message.API_ERROR, errors.Wrap(nil, "unknown API request type")
	}

}

// TODO: name should be changed
func (s *ServerCore) GetAllChainsForResolveConflict() error {
	log.Println("get all chains for resolve conflict called")
	return s.cm.SendMsgToAllPeer(nil, message.REQUEST_FULL_CHAIN)
}

func (s *ServerCore) VerifyTransaction(tx *transaction.Transaction, inBlock bool) bool {
	log.Println("VerifyTransaction is called")
	err := tx.HasValidSign()
	if err != nil {
		log.Printf("Invalid signiture : %v\n", err)
		return false
	}

	usedTxOutputs := tx.GetUsedOutputs()
	for _, txOut := range usedTxOutputs {
		if !inBlock && s.tp.HasTxOutput(&txOut) {
			log.Printf("txOut %v is already stored in TxPool", txOut)
			return false
		}
		if !s.bm.ValidateTxOut(&txOut) {
			return false
		}
	}
	return true
}
func (s *ServerCore) VerifyTransactionsInBlock(blk *block.Block) bool {
	log.Println("VerifyTransactionsInBlock is called")
	fee := blk.GetTotalFee()
	// TODO: insentive, but not here
	fee += 11
	numCoinbaseTX := 0
	for _, tx := range blk.Transactions {
		// TODO: check transactino type
		//switch tx := txI.(type) {
		if !tx.IsCoinBase { //case transaction.Transaction:
			if !s.VerifyTransaction(&tx, true) {
				log.Println("Bad transactions in block")
				return false
			}
		} else { //case transaction.CoinBaseTransaction:
			if numCoinbaseTX == 1 {
				log.Println("Single CoinbaseTransaction in one block is allowed")
				return false
			}
			insentive, _ := strconv.Atoi(tx.GetOutputs()[0].Value)
			if insentive != fee {
				log.Printf("Invalid fee in CoinbaseTransaction : expected(%d), actual(%d)\n", fee, insentive)
				return false
			}
			numCoinbaseTX++
		}
		//default:
		//panic("")
		//}
	}
	log.Println("VerifyTransactionsInBlock success")
	return true
}

func (s *ServerCore) handleMessage(msg *message.Message, peer *p2p.Node) error {
	log.Println("Server Core handleMessage called")

	if peer != nil {
		switch msg.Type {
		case message.REQUEST_FULL_CHAIN:
			log.Printf("Send latest blockchain for reply to : %v\n", peer)
			chain := s.bm.Chain
			err := s.SendTo(chain, message.RSP_FULL_CHAIN, peer)
			if err != nil {
				panic(err)
			}
		default:
			return errors.Wrap(nil, "Uknown message type")
		}
	}

	switch msg.Type {
	case message.NEW_TRANSACTION:
		var inTxs transaction.Transaction
		err := json.Unmarshal(msg.Payload, &inTxs)
		if err != nil {
			panic(err)
		}
		log.Printf("Received transactions : %s\n", inTxs)
		// TODO: is_sbc_transaction to be implemented by using interface's type assertion
		if s.HasSameTx(&inTxs) {
			log.Printf("transaction %s is already pooled", inTxs)
			return nil
		} else {
			// TODO: this should cause timing issue
			err = s.cm.SendMsgToAllPeer(msg.Payload, message.NEW_TRANSACTION)
			if err != nil {
				return err
			}
		}
		// TODO : this block is for debugging
		// if DEBUG {
		if len(s.bm.Chain) != 1 {
			if !s.VerifyTransaction(&inTxs, false) {
				log.Println("Transaction verification failed")
				return nil
			}
		}
		// }
		s.tp.Append(inTxs)

	case message.NEW_BLOCK:
		var newBlk block.Block
		err := json.Unmarshal([]byte(msg.Payload), &newBlk)
		if err != nil {
			panic(err)
		}
		log.Printf("New block : %v\n", newBlk)
		if s.bm.IsValidBlock([]byte(s.prvBlkHash), newBlk, block.DIFFICULTY) {
			log.Println("*************************************** Block Nonce Received *******************************************")
			// stop block generation
			s.genBlockLoopCancel()

			hashByte, err := s.bm.GetHash(&newBlk)
			if err != nil {
				panic(err)
			}
			s.prvBlkHash = hashByte
			s.bm.AppendNewBlock(&newBlk)
			//trimmedTxs := s.bm.RemoveDuplicateTransactions(newBlk.Transactions)
			//s.tp.OverwriteTransactions(trimmedTxs)
			s.tp.TrimTransactions(newBlk.Transactions)

			err = s.SendToAllEdges(newBlk, message.NEW_BLOCK)
			if err != nil {
				panic(err)
			}
			go s.generateBlockLoop()
		} else {
			err := s.GetAllChainsForResolveConflict()
			if err != nil {
				panic(err)
			}
		}
	case message.RSP_FULL_CHAIN:
		var commingChain block.BlockChain
		err := json.Unmarshal([]byte(msg.Payload), &commingChain)
		if err != nil {
			panic(err)
		}
		hash, resolvedBlocks := s.bm.ResolveBranch(commingChain)
		log.Printf("blockchain received : %s", commingChain)
		if hash != nil {
			s.prvBlkHash = hash
			if len(resolvedBlocks) != 0 {
				txs := s.bm.GetTransactionsFromOrphanBlocks(resolvedBlocks)
				s.tp.Append(txs...)
			}
		}
	case message.ENHANCED:
		log.Println("Received enhanced message")

		// TODO: dup check should not be here

		return s.mpmh.HandleMessage(msg, s.API)
	default:
		return errors.Wrap(nil, "Unknown message type")
	}
	return nil
}

func (s *ServerCore) HasSameTx(tx *transaction.Transaction) bool {
	return s.tp.Has(tx)
}
