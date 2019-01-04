package core

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p/message"
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
	State             CoreState
	IP                string
	Port              uint16 // can be conn struct?
	cm                p2p.ConnectionManagerI
	BootstrapCoreNode *p2p.Node
	mpmh              message.MessageHandler
	coreContext       context.Context // top level context
	coreCancel        context.CancelFunc
}

func newCore(port uint16, bootStrapNode *p2p.Node, apiCallback func(msg *message.Message, peer *p2p.Node) error, isCore bool) *Core {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Core{
		State:             STATE_INIT,
		Port:              port,
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
