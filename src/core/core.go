package core

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ami-GS/blockchainFromZero/src/block"
	"github.com/ami-GS/blockchainFromZero/src/key"
	"github.com/ami-GS/blockchainFromZero/src/key/utils"
	"github.com/ami-GS/blockchainFromZero/src/p2p"
	"github.com/ami-GS/blockchainFromZero/src/p2p/message"
	p2putils "github.com/ami-GS/blockchainFromZero/src/p2p/utils"
	"github.com/ami-GS/blockchainFromZero/src/transaction"
)

type CoreState uint16

const (
	STATE_INIT   CoreState = iota
	STATE_ACTIVE           // for client only?
	STATE_STANBY
	STATE_CONNECT_TO_NETWORK
	STATE_SHUTTING_DOWN
)

type CoreI interface {
	Start() (context.Context, context.CancelFunc)
	Shutdown()
	JoinNetwork(*p2p.Node) error
}

type Core struct {
	State CoreState
	IP    string
	Port  uint16 // can be conn struct?
	cm    p2p.ConnectionManagerI
	// TODO: BlockBuilder sould be in BlockChainManager
	bm                *block.BlockChainManager
	km                *key.KeyManager
	prvBlkHash        []byte
	BootstrapCoreNode *p2p.Node
	mpmh              *message.MessageHandler
	coreContext       context.Context // top level context
	coreCancel        context.CancelFunc
}

func newCore(port uint16, bootStrapNode *p2p.Node, apiCallback func(msg *message.Message, peer *p2p.Node) error, isCore bool) *Core {
	ctx, cancel := context.WithCancel(context.Background())
	bm := block.NewBlockChainManager()
	prvBlkHash, err := bm.Chain[0].GetHash() // genesis
	if err != nil {
		panic(err)
	}
	c := &Core{
		State:             STATE_INIT,
		Port:              port,
		bm:                bm,
		km:                key.New(),
		prvBlkHash:        prvBlkHash,
		BootstrapCoreNode: bootStrapNode,
		mpmh:              message.NewMessageHandler(),
		coreContext:       ctx,
		coreCancel:        cancel,
	}
	c.IP = strings.Split(p2putils.GetExternalIP(), "/")[0]
	if bootStrapNode != nil && c.Port == bootStrapNode.Port {
		// for local experiment purpose
		c.IP = "127.0.0.1"
		c.BootstrapCoreNode = nil
	}
	if isCore {
		log.Printf("Core address is %s:%d\n", c.IP, c.Port)
		c.cm = p2p.NewConnectionManagerCore(c.IP, port, bootStrapNode, apiCallback)
	} else {
		log.Printf("Edge address is %s:%d\n", c.IP, c.Port)
		c.cm = p2p.NewConnectionManager4Edge(c.IP, port, bootStrapNode, apiCallback)
	}
	return c
}

func (s *Core) jsonize(data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("failed to jsonize :", data)
		return nil, fmt.Errorf("failed to jsonize : %v", data)
	}
	return jsonData, nil
}

func (s *Core) Broadcast(data interface{}, msgType message.MessageType) error {
	jsonData, err := s.jsonize(data)
	if err != nil {
		return err
	}
	return s.cm.SendMsgBroadcast(jsonData, msgType)
}

func (s *Core) SendToAllCores(data interface{}, msgType message.MessageType) error {
	jsonData, err := s.jsonize(data)
	if err != nil {
		return err
	}
	return s.cm.SendMsgToAllPeer(jsonData, msgType)
}

func (s *Core) SendToAllEdges(data interface{}, msgType message.MessageType) error {
	jsonData, err := s.jsonize(data)
	if err != nil {
		return err
	}
	return s.cm.SendMsgToAllEdge(jsonData, msgType)
}

func (s *Core) SendTo(data interface{}, msgType message.MessageType, peer *p2p.Node) error {
	jsonData, err := s.jsonize(data)
	if err != nil {
		return err
	}
	return s.cm.SendMsgTo(jsonData, msgType, peer)
}

func (s *Core) Start() {
	s.State = STATE_STANBY
	s.cm.Start(s.coreContext)
}

func (s *Core) Shutdown() {
	s.State = STATE_SHUTTING_DOWN
	s.cm.ConnectionClose()
}

func (c *Core) GetTransactionsFromChain() []transaction.Transaction {
	return c.bm.GetTransactions()
}

func (c *Core) Sign(data []byte) ([]byte, error) {
	return c.km.Sign(data)
}

func (c *Core) RenewKey() {
	c.km = key.New()
}
func (c *Core) GetPrivateKey() *rsa.PrivateKey {
	return &c.km.PrivateKey
}
func (c *Core) GetPublicKey() *rsa.PublicKey {
	return &c.km.PrivateKey.PublicKey
}
func (c *Core) SetPrivateKey(key rsa.PrivateKey) {
	c.km.PrivateKey = key
}

func (c *Core) GetPublicKeyByHexString() string {
	return hex.EncodeToString(keyutils.PublicKeyToBytes(&c.km.PrivateKey.PublicKey))
}

func (c *Core) GetPublicKeyBytes() []byte {
	return keyutils.PublicKeyToBytes(&c.km.PrivateKey.PublicKey)
}

func (c *Core) GetPublicKeyBase64() []byte {
	return []byte(base64.StdEncoding.EncodeToString(keyutils.PublicKeyToBytes(&c.km.PrivateKey.PublicKey)))
}
