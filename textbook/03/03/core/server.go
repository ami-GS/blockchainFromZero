package core

import (
	"context"
	"log"
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/03/03/block"
	"github.com/ami-GS/blockchainFromZero/textbook/03/03/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/03/03/p2p/message"
	"github.com/ami-GS/blockchainFromZero/textbook/03/03/transaction"
	"github.com/pkg/errors"
)

type ServerCore struct {
	*Core
	protocolMessageStore map[string]struct{} // store payload, duplication check
	bb                   *block.BlockBuilder
	bm                   *block.BlockChainManager
	tp                   *transaction.TransactionPool
	prvBlkHash           string
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
		prvBlkHash:           string(prvBlkHash),
		tp:                   transaction.NewTransactionPool(),
	}
	s.Core = newCore(port, bootStrapNode, s.handleMessage, true)
	return s
}

func (s *ServerCore) Start() {
	s.State = STATE_STANBY
	ctx := s.Core.Start()
	go s.generateBlockLoop(ctx)
}

func (s *ServerCore) generateBlockLoop(ctx context.Context) {
	log.Printf("Generate block with transactions")
	genblock := func() {
		size, txs := s.tp.Get()
		if size == 0 {
			return
		}
		newBlk := s.bb.GenerateNewBlock(string(txs), s.prvBlkHash)
		s.bm.AppendNewBlock(newBlk)
		prvBlkHash, err := s.bm.GetHash(newBlk)
		if err != nil {
			panic(err)
		}
		s.prvBlkHash = string(prvBlkHash)
		s.tp.PopFront(size)
		log.Println("Current Blockchain is ...", s.bm.Chain)
		log.Println("Current prvBlkHash is ...", s.prvBlkHash)
	}

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		// case <- cancel
		// return
		case <-ticker.C:
			genblock()
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

func (s *ServerCore) handleMessage(msg *message.Message, peer *p2p.Node) error {
	log.Println("Server Core handleMessage called")
	switch msg.Type {
	case message.NEW_TRANSACTION:
	case message.NEW_BLOCK:
	case message.RSP_FULL_CHAIN:
	case message.ENHANCED:
		log.Println("Received enhanced message")
		// TODO: need mu.Lock()
		// TODO: use hash to check easily
		for payload, _ := range s.protocolMessageStore {
			if payload == msg.Payload {
				return nil
			}
		}
		s.protocolMessageStore[msg.Payload] = struct{}{}
		return s.mpmh.HandleMessage(msg, s.API)
	default:
		return errors.Wrap(nil, "Unknown message type")
	}
	return nil
}
