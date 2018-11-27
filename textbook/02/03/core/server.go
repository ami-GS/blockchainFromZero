package core

import (
	"log"

	"github.com/ami-GS/blockchainFromZero/textbook/02/03/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/02/03/p2p/message"
	"github.com/pkg/errors"
)

type ServerCore struct {
	*Core
	protocolMessageStore map[string]struct{} // store payload, duplication check
}

func NewServerCore(port uint16, bootStrapNode *p2p.Node) *ServerCore {
	log.Println("Initialize server ...")
	s := &ServerCore{
		protocolMessageStore: make(map[string]struct{}),
	}
	s.Core = newCore(port, bootStrapNode, s.handleMessage, true)
	return s
}

func (s *ServerCore) Start() {
	s.State = STATE_STANBY
	s.Core.Start()
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
