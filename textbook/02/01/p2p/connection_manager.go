package p2p

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const PING_INTERVAL = 10 //sec

type ConnectionManager struct {
	host          string
	port          uint16
	coreNodeSet   CoreNodeSet
	edgeNodeSet   EdgeNodeSet
	mm            *MessageManager
	bootstrapNode *Node
}

func NewConnectionManager(host string, port uint16, bootStrap *Node) *ConnectionManager {
	return &ConnectionManager{
		host:          host,
		port:          port,
		coreNodeSet:   CoreNodeSet{NewNodeSet()},
		edgeNodeSet:   EdgeNodeSet{NewNodeSet()},
		mm:            &MessageManager{},
		bootstrapNode: bootStrap,
	}
}

func (c *ConnectionManager) Start(ctx context.Context) {
	go c.waitForAccess(ctx)
	go func() {
		ticker := time.NewTicker(PING_INTERVAL * time.Second)
		for {
			select {
			case <-ticker.C:
				c.checkPeersConnection()
				// case <- :
				// TODO: cancel
			}
		}
	}()
}

func (c *ConnectionManager) JoinNetwork(bootstrapNode *Node) error {
	// for core node
	c.bootstrapNode = bootstrapNode
	return c.connectToP2PNW(bootstrapNode)
}

func (c *ConnectionManager) connectToP2PNW(bootstrapNode *Node) error {
	conn, err := net.Dial("tcp", bootstrapNode.String())
	defer conn.Close()
	if err != nil {
		return err
	}
	jsonMsg, err := c.mm.Build(ADD, c.port, "")
	if err != nil {
		return err
	}
	_, err = conn.Write(jsonMsg)
	return err
}

func (c *ConnectionManager) ConnectionClose() error {
	// conn, err := net.Dial("tcp", c.host+":"+strconv.Itoa(int(c.port)))
	// defer conn.Close()
	// if err != nil {
	// 	return err
	// }
	// TODO: send cancel to waitForAccess
	// TODO: send cancel to ticker.C of Start()
	// conn.Write()
	jsonMsg, err := c.mm.Build(REMOVE, c.port, "")
	if err != nil {
		return err
	}
	return c.sendMsg(c.bootstrapNode, jsonMsg)
}

func (c *ConnectionManager) handleMessage(conn net.Conn) error {
	b, err := ioutil.ReadAll(conn)

	if err != nil {
		log.Fatal(err)
	}

	rsp, msg, err := c.mm.Parse(b)
	if err != nil {
		panic(err)
	}
	if rsp == OK_WITHOUT_PAYLOAD {
		ipv4 := strings.Split(conn.RemoteAddr().String(), ":")
		switch msg.Type {
		case ADD:
			log.Println("ADD node request")
			c.addPeer(ipv4[0], msg.myPort)
			if ipv4[0] == c.host && msg.myPort == c.port {
				return nil
			}
			nodesStr := c.coreNodeSet.GetNodesByString()
			jsonMsg, err := c.mm.Build(CORE_LIST, c.port, nodesStr)
			if err != nil {
				panic(err)
			}
			c.sendMsgToAllPeer(jsonMsg)
			c.sendMsgToAllEdge(jsonMsg)
		case REMOVE:
			log.Printf("REMOVE node request from %d\n", ipv4[0], msg.myPort)
			c.removePeer(ipv4[0], msg.myPort)
			nodesStr := c.coreNodeSet.GetNodesByString()
			jsonMsg, err := c.mm.Build(CORE_LIST, c.port, nodesStr)
			if err != nil {
				panic(err)
			}
			c.sendMsgToAllPeer(jsonMsg)
			c.sendMsgToAllEdge(jsonMsg)
		case PING:
		case REQUEST_CORE_LIST:
			log.Println("List for Core nodes was requested")
			nodesStr := c.coreNodeSet.GetNodesByString()
			jsonMsg, err := c.mm.Build(CORE_LIST, c.port, nodesStr)
			if err != nil {
				panic(err)
			}
			c.sendMsg(&Node{ipv4[0], msg.myPort}, jsonMsg)
		default:
			panic(errors.Wrap(nil, "unknown command"))
		}
	} else if rsp == OK_WITH_PAYLOAD {
		switch msg.Type {
		case CORE_LIST:
			log.Printf("Refresh core node list ...")
		default:
			panic(errors.Wrap(nil, "unknown command"))
		}
	} else {
		panic(errors.Wrap(nil, "unknown response status"))
	}
	return nil
}

func (c *ConnectionManager) waitForAccess(ctx context.Context) {
	// TODO: pass ctx to handleMessage?
	listen, err := net.Listen("tcp", c.host+":"+strconv.Itoa(int(c.port)))
	if err != nil {
		panic(err)
	}
	defer listen.Close()

	for {
		log.Println("Waiting for the connection ...")
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Connected by .. %v\n", conn.RemoteAddr())
		go c.handleMessage(conn)
	}
}

func (c *ConnectionManager) addPeer(ip string, port uint16) {
	portStr := strconv.Itoa(int(port))
	c.coreNodeSet.Add(ip + ":" + portStr)
}

func (c *ConnectionManager) removePeer(ip string, port uint16) {
	portStr := strconv.Itoa(int(port))
	c.coreNodeSet.Remove(ip + ":" + portStr)
}

func (c *ConnectionManager) sendMsgToAllPeer(msg []byte) error {
	log.Println("Send message to all peer")
	nodes, err := c.coreNodeSet.GetNodes()
	if err != nil {
		return err
	}
	for _, peer := range nodes {
		if (*peer).IP != c.host && (*peer).port != c.port {
			c.sendMsg(peer, msg)
		}
	}
	return nil
}
func (c *ConnectionManager) sendMsgToAllEdge(msg []byte) error {
	log.Println("Send message to all edges")
	nodes, err := c.edgeNodeSet.GetNodes()
	if err != nil {
		return err
	}
	for _, edge := range nodes {
		c.sendMsg(edge, msg)
	}
	return nil
}

func (c *ConnectionManager) sendMsg(peer *Node, msg []byte) error {
	conn, err := net.Dial("tcp", peer.String())
	defer conn.Close()
	if err != nil {
		return err
	}
	_, err = conn.Write(msg)
	return nil
}

func (c *ConnectionManager) checkPeersConnection() {

}
