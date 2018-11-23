package core

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/ami-GS/blockchainFromZero/textbook/02/01/p2p"
)

type ServerState uint16

const (
	STATE_INIT ServerState = iota
	STATE_STANBY
	STATE_CONNECT_TO_CENTRAL
	STATE_SHUTTING_DOWN
)

type ServerCore struct {
	State             ServerState
	IP                string
	Port              uint16 // can be conn struct?
	cm                *p2p.ConnectionManager
	BootstrapCoreNode *p2p.Node
}

func NewServerCore(port uint16, bootStrapNode *p2p.Node) *ServerCore {
	log.Println("Initialize server ...")
	s := &ServerCore{
		State:             STATE_INIT,
		Port:              port,
		BootstrapCoreNode: bootStrapNode,
	}
	s.IP = strings.Split(s.GetMyExternalIP(), "/")[0]
	log.Printf("Server address is %s\n", s.IP)
	s.cm = p2p.NewConnectionManager(s.IP, port, bootStrapNode)
	return s
}

func (s *ServerCore) Start() {
	s.State = STATE_STANBY
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.cm.Start(ctx)
}

func (s *ServerCore) JoinNetwork() error {
	if s.BootstrapCoreNode == nil {
		log.Println("This server is running as genesis core node ...")
		return nil
	}
	s.State = STATE_CONNECT_TO_CENTRAL
	return s.cm.JoinNetwork(s.BootstrapCoreNode)
}

func (s *ServerCore) Shutdown() {
	s.State = STATE_SHUTTING_DOWN
	log.Println("Shutdown server ...")
	s.cm.ConnectionClose()
}

// TODO: in utils/ ?
func (s *ServerCore) GetMyExternalIP() string {
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
