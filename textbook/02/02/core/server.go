package core

import (
	"log"

	"github.com/ami-GS/blockchainFromZero/textbook/02/02/p2p"
)

type ServerCore struct {
	*Core
}

func NewServerCore(port uint16, bootStrapNode *p2p.Node) *ServerCore {
	log.Println("Initialize server ...")
	return &ServerCore{
		newCore(port, bootStrapNode, true),
	}
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
