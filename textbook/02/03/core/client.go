package core

import (
	"log"

	"github.com/ami-GS/blockchainFromZero/textbook/02/03/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/02/03/p2p/message"
	"github.com/pkg/errors"
)

type ClientCore struct {
	*Core
}

func NewClientCore(port uint16, bootStrapNode *p2p.Node) *ClientCore {
	log.Println("Initialize edge ...")
	c := &ClientCore{}
	c.Core = newCore(port, bootStrapNode, c.handleMessage, false)
	return c
}

func (c *ClientCore) Start() {
	c.State = STATE_ACTIVE
	c.Core.Start()
	c.JoinNetwork()
}

func (c *ClientCore) SendMessageToMyCore(msgType message.MessageType, msg []byte) error {
	// TODO: implement method to retrieve actual message
	log.Println("Send Message:", string(msg))
	jsonMsg, err := c.cm.GetMessage(msgType, msg)
	if err != nil {
		return err
	}
	return c.cm.SendMsg(c.BootstrapCoreNode, jsonMsg)
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
	log.Println("Client handleMessage called")
	switch msg.Type {
	case message.RSP_FULL_CHAIN:
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
