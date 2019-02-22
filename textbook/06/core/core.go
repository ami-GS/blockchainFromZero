package core

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/ami-GS/blockchainFromZero/textbook/06/block"
	"github.com/ami-GS/blockchainFromZero/textbook/06/key"
	"github.com/ami-GS/blockchainFromZero/textbook/06/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/06/p2p/message"
	"github.com/ami-GS/blockchainFromZero/textbook/06/transaction"
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
	bb                *block.BlockBuilder
	bm                *block.BlockChainManager
	km                *key.KeyManager
	prvBlkHash        []byte
	BootstrapCoreNode *p2p.Node
	mpmh              message.MessageHandler
	coreContext       context.Context // top level context
	coreCancel        context.CancelFunc
}

func newCore(port uint16, bootStrapNode *p2p.Node, apiCallback func(msg *message.Message, peer *p2p.Node) error, isCore bool) *Core {
	ctx, cancel := context.WithCancel(context.Background())
	bb := block.NewBlockBuilder()
	genesis := bb.GenerateGenesisBlock()
	bm := block.NewBlockChainManager(genesis)
	prvBlkHash, _ := bm.GetHash(((*block.Block)(genesis)))
	c := &Core{
		State:             STATE_INIT,
		Port:              port,
		bb:                bb,
		bm:                bm,
		km:                key.New(),
		prvBlkHash:        prvBlkHash,
		BootstrapCoreNode: bootStrapNode,
		mpmh:              message.MessageHandler{},
		coreContext:       ctx,
		coreCancel:        cancel,
	}
	c.IP = strings.Split(c.GetMyExternalIP(), "/")[0]
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

// TODO: in utils/ ?
func (s *Core) GetMyExternalIP() string {
	debugVal := os.Getenv("DEBUG_LOCAL_IP")
	if debugVal != "" {
		return "127.0.0.1"
	}

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			tmp := addr.String()
			if strings.Count(tmp, ".") == 3 && !strings.HasPrefix(tmp, "127.0.0.1") {
				return strings.Split(tmp, "/")[0]
			}
		}
	}
	return ""

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
func (c *Core) SetPrivateKey(key rsa.PrivateKey) {
	c.km.PrivateKey = key
}

func (c *Core) GetPublicKeyByHexString() string {
	return c.km.GetAddressByHexString()
}

func (c *Core) GetPublicKeyBytes() []byte {
	return c.km.PrivateKey.PublicKey.N.Bytes()
}

func (c *Core) GetPublicKeyBase64() []byte {
	return []byte(base64.StdEncoding.EncodeToString(c.km.PrivateKey.PublicKey.N.Bytes()))
}
