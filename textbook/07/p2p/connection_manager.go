package p2p

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	. "github.com/ami-GS/blockchainFromZero/textbook/07/p2p/message"
)

const PING_INTERVAL = 20 //sec

type ConnectionManagerI interface {
	Start(context.Context)
	JoinNetwork(*Node, []byte) error
	GetMessageBytes(MessageType, []byte) ([]byte, error)
	connectToP2PNW(*Node) error
	ConnectionClose() error
	handleMessage(net.Conn) error
	waitForAccess(context.Context)
	sendMsg(*Node, []byte) error
	SendMsgTo([]byte, MessageType, *Node) error
	SendMsgToAllPeer([]byte, MessageType) error
	SendMsgToAllEdge([]byte, MessageType) error
	SendMsgBroadcast([]byte, MessageType) error

	HasEdgeByPubkey(string) bool
	GetEdgeByPubkey(string) (string, int, error)
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
	// WARN: only for client for now
	pubkey []byte
}

func NewConnectionManager(host string, port uint16, bootStrap *Node, callback func(*Message, *Node) error, isCore bool) *ConnectionManager {
	coreNodeSet := CoreNodeSet{NewNodeSet()}
	coreNodeSet.AddString(fmt.Sprintf("%s:%d", host, port))
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

func (c *ConnectionManager) JoinNetwork(bootstrapNode *Node, pubkey []byte) error {
	// for core node
	c.bootstrapNode = bootstrapNode
	// TODO: refactoring this
	if !c.isCore {
		c.pubkey = pubkey
	}
	return c.connectToP2PNW(bootstrapNode)
}

// TODO: can be implement in each child
func (c *ConnectionManager) connectToP2PNW(bootstrapNode *Node) error {
	debugVal := os.Getenv("DEBUG_LOCAL_IP")
	if debugVal != "" {
		bootstrapNode.IP = "127.0.0.1"
	}

	conn, err := net.Dial("tcp", bootstrapNode.String())
	if err != nil {
		return err
	}
	defer conn.Close()
	var jsonMsg []byte
	if c.isCore {
		jsonMsg, err = c.mm.Build(ADD, c.port, nil)
	} else {
		jsonMsg, err = c.mm.Build(ADD_AS_EDGE, c.port, c.pubkey)
	}
	if err != nil {
		return err
	}
	_, msgLog, _ := c.mm.Parse(jsonMsg)
	log.Println("send message:1", msgLog)
	_, err = conn.Write(jsonMsg)
	return err
}

func (c *ConnectionManager) GetMessageBytes(msgType MessageType, msg []byte) ([]byte, error) {
	// TODO: if can be removed
	if msg == nil {
		return c.mm.Build(msgType, c.port, nil)
	}
	return c.mm.Build(msgType, c.port, msg)
}

func (c *ConnectionManager) SendMsgToAllPeer(msg []byte, mType MessageType) error {
	log.Println("WARN: SendMsgToAllPeer is called from client")
	return nil
}
func (c *ConnectionManager) SendMsgToAllEdge(msg []byte, mType MessageType) error {
	log.Println("WARN: SendMsgToAllEdge is called from client")
	return nil
}
func (c *ConnectionManager) SendMsgBroadcast(msg []byte, mType MessageType) error {
	log.Println("WARN: SendMsgBroadcast is called from client")
	return nil
}

func (c *ConnectionManager) HasEdgeByPubkey(pubkey string) bool {
	// search base64
	return c.edgeNodeSet.HasByPubkey(pubkey)
}

func (c *ConnectionManager) GetEdgeByPubkey(pubkey string) (string, int, error) {
	return c.edgeNodeSet.GetByPubkey(pubkey)
}
