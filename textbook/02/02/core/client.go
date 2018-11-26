package core

import (
	"log"

	"github.com/ami-GS/blockchainFromZero/textbook/02/02/p2p"
)

type ClientCore struct {
	*Core
}

func NewClientCore(port uint16, bootStrapNode *p2p.Node) *ClientCore {
	log.Println("Initialize edge ...")
	return &ClientCore{
		newCore(port, bootStrapNode, false),
	}
}

func (c *ClientCore) Start() {
	c.State = STATE_ACTIVE
	c.Core.Start()
	c.JoinNetwork()
}

func (c *ClientCore) JoinNetwork() error {
	c.State = STATE_CONNECT_TO_NETWORK
	return c.cm.JoinNetwork(c.BootstrapCoreNode)
}

func (c *ClientCore) Shutdown() {
	log.Println("Shutdown client ...")
	c.Core.Shutdown()
}
