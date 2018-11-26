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

type ConnectionManagerI interface {
	Start(context.Context)
	JoinNetwork(*Node) error
	connectToP2PNW(*Node) error
	ConnectionClose() error
	handleMessage(net.Conn) error
	waitForAccess(context.Context)
	sendMsg(*Node, []byte) error
}

type ConnectionManager struct {
	host          string
	port          uint16
	coreNodeSet   CoreNodeSet
	edgeNodeSet   EdgeNodeSet
	mm            *MessageManager
	bootstrapNode *Node
}

func NewConnectionManager(host string, port uint16, bootStrap *Node) ConnectionManagerI {
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
		pingToPeer := time.NewTicker(PING_INTERVAL * time.Second)
		pingToEdge := time.NewTicker(PING_INTERVAL * time.Second)
		for {
			select {
			case <-pingToPeer.C:
				c.checkPeersConnection()
			case <-pingToEdge.C:
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
	// TODO: send cancel to waitForAccess
	// TODO: send cancel to pingToPeer.C of Start()
	// TODO: send cancel to pingToEdge.C of Start()
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
		panic(err)
	}

	rsp, msg, err := c.mm.Parse(b)
	if err != nil {
		panic(err)
	}
	if rsp == OK_WITHOUT_PAYLOAD {
		ipv4 := strings.Split(conn.RemoteAddr().String(), ":")
		switch msg.Type {
		case ADD, ADD_AS_EDGE:
			c.addPeer(ipv4[0], msg.MyPort, msg.Type)
			if ipv4[0] == c.host && msg.MyPort == c.port {
				return nil
			}
			nodesStr := c.coreNodeSet.GetNodesByString()
			jsonMsg, err := c.mm.Build(CORE_LIST, c.port, nodesStr)
			if err != nil {
				panic(err)
			}
			if msg.Type == ADD {
				log.Println("ADD node request")
				c.sendMsgToAllPeer(jsonMsg)
				c.sendMsgToAllEdge(jsonMsg)
			} else {
				log.Println("ADD_AS_EDGE request")
				c.sendMsg(&Node{ipv4[0], msg.MyPort}, jsonMsg)
			}
		case REMOVE:
			log.Printf("REMOVE request from %d\n", ipv4[0], msg.MyPort)
			c.removePeer(ipv4[0], msg.MyPort, msg.Type)
			nodesStr := c.coreNodeSet.GetNodesByString()
			jsonMsg, err := c.mm.Build(CORE_LIST, c.port, nodesStr)
			if err != nil {
				panic(err)
			}
			c.sendMsgToAllPeer(jsonMsg)
			c.sendMsgToAllEdge(jsonMsg)
		case REMOVE_EDGE:
			log.Printf("REMOVE_EDGE request from %d\n", ipv4[0], msg.MyPort)
			c.removePeer(ipv4[0], msg.MyPort, msg.Type)
		case PING:
		case REQUEST_CORE_LIST:
			log.Println("List for Core nodes was requested")
			nodesStr := c.coreNodeSet.GetNodesByString()
			jsonMsg, err := c.mm.Build(CORE_LIST, c.port, nodesStr)
			if err != nil {
				panic(err)
			}
			c.sendMsg(&Node{ipv4[0], msg.MyPort}, jsonMsg)
		default:
			panic(errors.Wrap(nil, "unknown command"))
		}
	} else if rsp == OK_WITH_PAYLOAD {
		switch msg.Type {
		case CORE_LIST:
			log.Println("Refresh core node list ...")
			c.coreNodeSet.OverWriteByString(msg.Payload)
			log.Println("latest core node list:", c.coreNodeSet.GetNodesByString())
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

func (c *ConnectionManager) addPeer(ip string, port uint16, msType MessageType) {
	portStr := strconv.Itoa(int(port))
	if msType == ADD {
		log.Printf("Adding Core: %s:%s", ip, portStr)
		c.coreNodeSet.Add(ip + ":" + portStr)
	} else if msType == ADD_AS_EDGE {
		log.Printf("Adding Edge: %s:%s", ip, portStr)
		c.edgeNodeSet.Add(ip + ":" + portStr)
	}
}

func (c *ConnectionManager) removePeer(ip string, port uint16, msType MessageType) {
	portStr := strconv.Itoa(int(port))
	if msType == REMOVE {
		log.Printf("Removing Core: %s:%s", ip, portStr)
		c.coreNodeSet.Remove(ip + ":" + portStr)
	} else if msType == REMOVE_EDGE {
		c.edgeNodeSet.Remove(ip + ":" + portStr)
	}
}

func (c *ConnectionManager) sendMsgToAllPeer(msg []byte) error {
	log.Println("Send message to all peer")
	nodes, err := c.coreNodeSet.GetNodes()
	if err != nil {
		return err
	}

	for _, peer := range nodes {
		if (*peer).IP != c.host || (*peer).port != c.port {
			err = c.sendMsg(peer, msg)
			if err != nil {
				return err
				// TODO: continue with disconnection?
			}
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
		err = c.sendMsg(edge, msg)
		if err != nil {
			return err
			// TODO: continue with disconnection?
		}
	}
	return nil
}

func (c *ConnectionManager) sendMsg(peer *Node, msg []byte) error {
	conn, err := net.Dial("tcp", peer.String())
	defer conn.Close()
	if err != nil {
		return err
	}
	log.Println("Send msg:", string(msg))
	_, err = conn.Write(msg)
	return err
}

func (c *ConnectionManager) checkPeersConnection() error {
	log.Println("check peer connections")
	currentCoreList, err := c.coreNodeSet.GetNodes()
	if err != nil {
		return err
	}
	outNodes := NewNodeSet()
	// TODO: can be multi thread
	for _, n := range currentCoreList {
		err = c.ping(n)
		if err != nil {
			// TODO: need to check error type
			outNodes.Add(n.String())
		}
	}

	c.coreNodeSet.Sub(outNodes)
	currentCoreListStr := c.coreNodeSet.GetNodesByString()
	if err != nil {
		return err
	}
	log.Println("current core nodes are:", currentCoreListStr)
	if outNodes.Len() > 0 {
		jsonMsg, err := c.mm.Build(CORE_LIST, c.port, currentCoreListStr)
		if err != nil {
			panic(err)
		}
		c.sendMsgToAllEdge(jsonMsg)
		c.sendMsgToAllPeer(jsonMsg)
	}
	return nil
}

func (c *ConnectionManager) checkEdgesConnection() error {
	log.Println("check edge connections")
	currentEdgeList, err := c.edgeNodeSet.GetNodes()
	if err != nil {
		return err
	}
	outNodes := NewNodeSet()
	// TODO: can be multi thread
	for _, n := range currentEdgeList {
		err = c.ping(n)
		if err != nil {
			// TODO: need to check error type
			outNodes.Add(n.String())
		}
	}

	c.edgeNodeSet.Sub(outNodes)
	currentEdgeListStr := c.edgeNodeSet.GetNodesByString()
	if err != nil {
		return err
	}
	log.Println("current edge nodes are:", currentEdgeListStr)
	return nil
}

func (c *ConnectionManager) ping(target *Node) error {
	conn, err := net.Dial("tcp", target.String())
	defer conn.Close()
	// TODO: error handling
	if err != nil {
		return err
	}
	// TODO: can be constant
	jsonMsg, _ := c.mm.Build(PING, c.port, "")

	_, err = conn.Write(jsonMsg)
	if err != nil {
		return err
	}
	return nil
}
