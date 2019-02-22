package core

import (
	"context"
	"crypto/aes"
	"encoding/json"
	"log"

	"github.com/ami-GS/blockchainFromZero/src/block"
	"github.com/ami-GS/blockchainFromZero/src/key/utils"
	"github.com/ami-GS/blockchainFromZero/src/p2p"
	"github.com/ami-GS/blockchainFromZero/src/p2p/message"
	"github.com/pkg/errors"
)

type ClientCore struct {
	*Core
	updateChainCallback func()
	dmReceivedCallback  func(map[string]string)
}

func NewClientCore(port uint16, bootStrapNode *p2p.Node) *ClientCore {
	log.Println("Initialize edge ...")
	c := &ClientCore{}
	c.Core = newCore(port, bootStrapNode, c.handleMessage, false)
	return c
}

func (c *ClientCore) Start(pubkey []byte) (context.Context, context.CancelFunc) {
	c.State = STATE_ACTIVE
	c.Core.Start()
	c.JoinNetwork(pubkey)
	return c.coreContext, c.coreCancel
}

func (c *ClientCore) SendMessageToMyCore(msgType message.MessageType, msg []byte) error {
	// TODO: implement method to retrieve actual message
	log.Println("Send Message:", string(msg))
	return c.cm.SendMsgTo(msg, msgType, c.BootstrapCoreNode)
}

func (c *ClientCore) JoinNetwork(pubkey []byte) error {
	c.State = STATE_CONNECT_TO_NETWORK
	return c.cm.JoinNetwork(c.BootstrapCoreNode, pubkey)
}

func (c *ClientCore) API(request message.RequestType, msg *message.Message) (message.APIResponseType, error) {
	switch request {
	case message.GET_API_ORIGIN:
		return message.CLIENT_CORE_API, nil
	case message.PASS_TO_CLIENT_API:
		enhancedMsg := make(map[string]string)
		json.Unmarshal(msg.Payload, &enhancedMsg)
		if cipherAesKeyBase64, ok := enhancedMsg["Key"]; ok {
			// TODO: make clean
			cipherAesKey, err := keyutils.DecodeBase64([]byte(cipherAesKeyBase64))
			// WA, 2 bytes long for some reason
			cipherAesKey = cipherAesKey[:len(cipherAesKey)-2]
			if err != nil {
				return message.API_ERROR, err
			}
			privKey := c.GetPrivateKey()
			aesKey := keyutils.DecryptWithPrivateKey(cipherAesKey, privKey)
			cipherMsgBase64 := enhancedMsg["Body"]
			cipherMsg, err := keyutils.DecodeBase64([]byte(cipherMsgBase64))
			if err != nil {
				return message.API_ERROR, err
			}
			blk, err := aes.NewCipher(aesKey)
			if err != nil {
				panic(err)
			}

			plaintext := keyutils.DecryptSplit(cipherMsg, blk, 16)

			enhancedMsg["Key"] = string(aesKey)
			enhancedMsg["Body"] = string(plaintext)
		}

		c.dmReceivedCallback(enhancedMsg)
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

func (c *ClientCore) SetDMReceivedCallback(callback func(msg map[string]string)) {
	log.Println("SetDMReceivedCallback is called")
	c.dmReceivedCallback = callback
}

func (c *ClientCore) GetFullCain() error {
	return c.SendMessageToMyCore(message.REQUEST_FULL_CHAIN, nil)
}
