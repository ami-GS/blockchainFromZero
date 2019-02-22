package core

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ami-GS/blockchainFromZero/textbook/05/block"
	"github.com/ami-GS/blockchainFromZero/textbook/05/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/05/p2p/message"
	"github.com/pkg/errors"
)

type ClientCore struct {
	*Core
	updateChainCallback func()
}

func NewClientCore(port uint16, bootStrapNode *p2p.Node) *ClientCore {
	log.Println("Initialize edge ...")
	c := &ClientCore{}
	c.Core = newCore(port, bootStrapNode, c.handleMessage, false)
	return c
}

func (c *ClientCore) Start() (context.Context, context.CancelFunc) {
	c.State = STATE_ACTIVE
	c.Core.Start()
	c.JoinNetwork()
	return c.coreContext, c.coreCancel
}

func (c *ClientCore) SendMessageToMyCore(msgType message.MessageType, msg []byte) error {
	// TODO: implement method to retrieve actual message
	log.Println("Send Message:", string(msg))
	return c.cm.SendMsgTo(msg, msgType, c.BootstrapCoreNode)
}

func (c *ClientCore) JoinNetwork() error {
	c.State = STATE_CONNECT_TO_NETWORK
	return c.cm.JoinNetwork(c.BootstrapCoreNode)
}

func (c *ClientCore) API(request message.RequestType, msg *message.Message) (message.APIResponseType, error) {
	switch request {
	case message.GET_API_ORIGIN:
		return message.CLIENT_CORE_API, nil
	case message.PASS_TO_CLIENT_API:
		return message.API_OK, nil
	default:
		return message.API_ERROR, errors.Wrap(nil, "unknown API request type")
	}
}

func (c *ClientCore) handleMessage(msg *message.Message, peer *p2p.Node) error {
	// TODO: noncence arcitecture, to be moved to p2p client core
	log.Println("Client handleMessage called")
	switch msg.Type {
	case message.RSP_FULL_CHAIN:
		var chain block.BlockChain
		err := json.Unmarshal(msg.Payload, &chain)
		if err != nil {
			return err
		}
		hash, _ := c.bm.ResolveBranch(chain)
		if hash != nil {
			c.prvBlkHash = hash
			c.updateChainCallback()
		}
	case message.NEW_BLOCK:
		c.GetFullCain()
	case message.ENHANCED:
		log.Println("Received enhanced message")
		return c.mpmh.HandleMessage(msg, c.API)
	default:
		return errors.Wrap(nil, "Unknown message type")
	}
	return nil
}

func (c *ClientCore) Shutdown() {
	log.Println("Shutdown client ...")
	c.Core.Shutdown()
}

func (c *ClientCore) SetBlockchainUpdateCallback(callback func()) {
	log.Println("SetBlockchainUpdateCallback is called")
	c.updateChainCallback = callback
}

func (c *ClientCore) GetFullCain() error {
	return c.SendMessageToMyCore(message.REQUEST_FULL_CHAIN, nil)
}
