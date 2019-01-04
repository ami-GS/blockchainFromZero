package p2p

import (
	"context"
	"fmt"
	"log"
	"net"

	. "github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p/message"
)

const PING_INTERVAL = 20 //sec

type ConnectionManagerI interface {
	Start(context.Context)
	JoinNetwork(*Node) error
	GetMessageBytes(MessageType, []byte) ([]byte, error)
	connectToP2PNW(*Node) error
	ConnectionClose() error
	handleMessage(net.Conn) error
	waitForAccess(context.Context)
	SendMsg(*Node, []byte) error
	SendMsgToAllPeer([]byte) error
	SendMsgToAllEdge([]byte) error
}

type ConnectionManager struct {
	host          string
	port          uint16
	coreNodeSet   CoreNodeSet
	edgeNodeSet   EdgeNodeSet
	mm            *MessageManager
	bootstrapNode *Node
	isCore        bool
	callback      func(*Message, *Node) error
}

func NewConnectionManager(host string, port uint16, bootStrap *Node, callback func(*Message, *Node) error, isCore bool) *ConnectionManager {
	coreNodeSet := CoreNodeSet{NewNodeSet()}
	coreNodeSet.Add(fmt.Sprintf("%s:%d", host, port))
	return &ConnectionManager{
		host:          host,
		port:          port,
		coreNodeSet:   coreNodeSet,
		edgeNodeSet:   EdgeNodeSet{NewNodeSet()},
		mm:            &MessageManager{},
		bootstrapNode: bootStrap,
		isCore:        isCore,
		callback:      callback,
	}
}

func (c *ConnectionManager) JoinNetwork(bootstrapNode *Node) error {
	// for core node
	c.bootstrapNode = bootstrapNode
	return c.connectToP2PNW(bootstrapNode)
}

func (c *ConnectionManager) connectToP2PNW(bootstrapNode *Node) error {
	conn, err := net.Dial("tcp", bootstrapNode.String())
	if err != nil {
		return err
	}
	defer conn.Close()
	var jsonMsg []byte
	if c.isCore {
		jsonMsg, err = c.mm.Build(ADD, c.port, nil)
	} else {
		jsonMsg, err = c.mm.Build(ADD_AS_EDGE, c.port, nil)
	}
	if err != nil {
		return err
	}
	_, msgLog, _ := c.mm.Parse(jsonMsg)
	log.Println("send message:", msgLog)
	_, err = conn.Write(jsonMsg)
	return err
}

func (c *ConnectionManager) GetMessageBytes(msgType MessageType, msg []byte) ([]byte, error) {
	if msg == nil {
		return c.mm.Build(msgType, c.port, nil)
	}
	return c.mm.Build(msgType, c.port, msg)
}

func (c *ConnectionManager) SendMsgToAllPeer(msg []byte) error {
	log.Println("WARN: SendMsgToAllPeer is called from client")
	return nil
}
func (c *ConnectionManager) SendMsgToAllEdge(msg []byte) error {
	log.Println("WARN: SendMsgToAllEdge is called from client")
	return nil
}

// TODO: ConnectionManager should have MessageHandler interface
// func (c *ConnectionManager) waitForAccess(ctx context.Context) {
// 	// TODO: pass ctx to handleMessage?
// 	listen, err := net.Listen("tcp", c.host+":"+strconv.Itoa(int(c.port)))
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer listen.Close()

// 	for {
// 		log.Println("Waiting for the connection ...")
// 		conn, err := listen.Accept()
// 		defer conn.Close()
// 		if err != nil {
// 			panic(err)
// 		}
// 		log.Printf("Connected by .. %v\n", conn.RemoteAddr())
// 		go c.handleMessage(conn)
// 	}
// }
