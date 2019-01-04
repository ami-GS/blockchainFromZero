package core

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/03/04/block"
	"github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p/message"
	"github.com/ami-GS/blockchainFromZero/textbook/03/04/transaction"
	"github.com/pkg/errors"
)

type ServerCore struct {
	*Core
	protocolMessageStore map[string]struct{} // store payload, duplication check
	bb                   *block.BlockBuilder
	bm                   *block.BlockChainManager
	tp                   *transaction.TransactionPool
	prvBlkHash           []byte
	genBlockLoopCancel   context.CancelFunc
	genBlockInterval     time.Duration
}

func NewServerCore(port uint16, bootStrapNode *p2p.Node) *ServerCore {
	log.Println("Initialize server ...")
	bb := block.NewBlockBuilder()
	genesis := bb.GenerateGenesisBlock()
	bm := block.NewBlockChainManager(genesis)
	prvBlkHash, _ := bm.GetHash(((*block.Block)(genesis)))
	s := &ServerCore{
		protocolMessageStore: make(map[string]struct{}),
		bb:                   bb,
		bm:                   bm,
		prvBlkHash:           prvBlkHash,
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
		if txs == nil {
			return nil
		}
		// TODO: suspicious ->
		trimmedTxs := s.bm.RemoveDuplicateTransactions(txs)
		s.tp.OverwriteTransactions(trimmedTxs)
		if len(trimmedTxs) == 0 {
			return nil
		}
		// <-

		// WARN: this is stopped if valid NEW_BLOCK comes
		newBlk := s.bb.GenerateNewBlock(trimmedTxs, s.prvBlkHash, &thisCtx)
		if newBlk == nil {
			// nil means process canceled
			return nil
		}

		s.bm.AppendNewBlock(newBlk)
		prvBlkHash, err := s.bm.GetHash(newBlk)
		if err != nil {
			panic(err)
		}
		s.prvBlkHash = prvBlkHash
		s.tp.PopFront(len(trimmedTxs))
		logPrintBlockInfo()
		return newBlk
	}
	genMsgAndSend := func(blk *block.Block) {
		if blk == nil {
			return
		}
		jsonBlk, err := json.Marshal(blk)
		if err != nil {
			log.Println("failed to jsonize new block :", blk)
			return
		}
		jsonMsg, err := s.cm.GetMessageBytes(message.NEW_BLOCK, jsonBlk)
		if err != nil {
			log.Println("failed to generate NEW_BLOCK message :", string(jsonBlk))
			return
		}
		err = s.cm.SendMsgToAllPeer(jsonMsg)
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
	return s.cm.JoinNetwork(s.BootstrapCoreNode)
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
		return message.API_OK, nil
	case message.SEND_TO_ALL_EDGE:
		return message.API_OK, nil
	default:
		return message.API_ERROR, errors.Wrap(nil, "unknown API request type")
	}
}

// TODO: name should be changed
func (s *ServerCore) GetAllChainsForResolveConflict() error {
	log.Println("get all chains for resolve conflict called")
	msg, err := s.cm.GetMessageBytes(message.REQUEST_FULL_CHAIN, nil)
	if err != nil {
		return err
	}
	err = s.Core.cm.SendMsgToAllPeer(msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServerCore) handleMessage(msg *message.Message, peer *p2p.Node) error {
	log.Println("Server Core handleMessage called")

	if peer != nil {
		switch msg.Type {
		case message.REQUEST_FULL_CHAIN:
			log.Printf("Send latest blockchain for reply to : %v\n", peer)
			chain := s.bm.Chain
			chainBytes, err := json.Marshal(chain)
			if err != nil {
				panic(err)
			}
			outMsg, err := s.cm.GetMessageBytes(message.RSP_FULL_CHAIN, chainBytes)
			if err != nil {
				panic(err)
			}
			err = s.cm.SendMsg(peer, outMsg)
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
		err := json.Unmarshal([]byte(msg.Payload), &inTxs)
		if err != nil {
			panic(err)
		}
		log.Printf("Received transactions : %s\n", inTxs)
		if s.tp.Has(inTxs) {
			log.Printf("transaction %s is already pooled", inTxs)
			return nil
		}
		s.tp.Append(inTxs)
	case message.NEW_BLOCK:
		var newBlk block.Block
		err := json.Unmarshal([]byte(msg.Payload), &newBlk)
		if err != nil {
			panic(err)
		}
		log.Printf("New block : %v\n", newBlk)
		if s.bm.IsValidBlock([]byte(s.prvBlkHash), newBlk, block.DIFFICULTY) {
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
		return s.mpmh.HandleMessage(msg, s.API)
	default:
		return errors.Wrap(nil, "Unknown message type")
	}
	return nil
}
