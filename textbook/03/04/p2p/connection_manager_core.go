package p2p

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	. "github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p/message"
	"github.com/pkg/errors"
)

type ConnectionManagerCore struct {
	*ConnectionManager
}

func NewConnectionManagerCore(host string, port uint16, bootstrap *Node, apiCallback func(*Message, *Node) error) ConnectionManagerI {
	return &ConnectionManagerCore{
		ConnectionManager: NewConnectionManager(host, port, bootstrap, apiCallback, true),
	}
}

func (c *ConnectionManagerCore) Start(ctx context.Context) {
	go c.waitForAccess(ctx)
	go func() {
		pingToPeer := time.NewTicker(PING_INTERVAL * time.Second)
		pingToEdge := time.NewTicker(PING_INTERVAL * time.Second)
		for {
			select {
			case <-pingToPeer.C:
				c.checkPeersConnection()
			case <-pingToEdge.C:
				c.checkEdgesConnection()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *ConnectionManagerCore) waitForAccess(ctx context.Context) {
	// TODO: pass ctx to handleMessage?
	listen, err := net.Listen("tcp", c.host+":"+strconv.Itoa(int(c.port)))
	if err != nil {
		panic(err)
	}
	defer listen.Close()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("=======================================================")
			return
		default:
			log.Println("Waiting for the connection ...")
			conn, err := listen.Accept()
			if err != nil {
				panic(err)
			}
			log.Printf("Connected by .. %v\n", conn.RemoteAddr())
			go c.handleMessage(conn)
		}
	}
}

func (c *ConnectionManagerCore) ConnectionClose() error {
	// TODO: send cancel to waitForAccess
	// TODO: send cancel to pingToPeer.C of Start()
	// TODO: send cancel to pingToEdge.C of Start()
	// conn.Write()
	jsonMsg, err := c.mm.Build(REMOVE, c.port, nil)
	if err != nil {
		return err
	}
	return c.SendMsg(c.bootstrapNode, jsonMsg)
}

func (c *ConnectionManagerCore) handleMessage(conn net.Conn) error {
	defer conn.Close()
	b, err := ioutil.ReadAll(conn)

	if err != nil {
		panic(err)
	}

	rsp, msg, err := c.mm.Parse(b)
	if err != nil {
		panic(err)
	}
	log.Println("handle message:", rsp, msg)
	if rsp == OK_WITHOUT_PAYLOAD {
		ipv4 := strings.Split(conn.RemoteAddr().String(), ":")
		switch msg.Type {
		case ADD, ADD_AS_EDGE:
			c.addPeer(ipv4[0], msg.MyPort, msg.Type)
			if ipv4[0] == c.host && msg.MyPort == c.port {
				return nil
			}
			nodesStr := c.coreNodeSet.GetNodesByString()
			if nodesStr == "" {
				return nil
			}
			jsonMsg, err := c.mm.Build(CORE_LIST, c.port, []byte(nodesStr))
			if err != nil {
				panic(err)
			}
			if msg.Type == ADD {
				log.Println("ADD node request")
				c.SendMsgToAllPeer(jsonMsg)
				c.SendMsgToAllEdge(jsonMsg)
			} else {
				log.Println("ADD_AS_EDGE request")
				c.SendMsg(&Node{ipv4[0], msg.MyPort}, jsonMsg)
			}
		case REMOVE:
			log.Printf("REMOVE request from %d\n", ipv4[0], msg.MyPort)
			c.removePeer(ipv4[0], msg.MyPort, msg.Type)
			nodesStr := c.coreNodeSet.GetNodesByString()
			jsonMsg, err := c.mm.Build(CORE_LIST, c.port, []byte(nodesStr))
			if err != nil {
				panic(err)
			}
			c.SendMsgToAllPeer(jsonMsg)
			c.SendMsgToAllEdge(jsonMsg)
		case REMOVE_EDGE:
			log.Printf("REMOVE_EDGE request from %d\n", ipv4[0], msg.MyPort)
			c.removePeer(ipv4[0], msg.MyPort, msg.Type)
		case PING:
		case REQUEST_CORE_LIST:
			log.Println("List for Core nodes was requested")
			nodesStr := c.coreNodeSet.GetNodesByString()
			jsonMsg, err := c.mm.Build(CORE_LIST, c.port, []byte(nodesStr))
			if err != nil {
				panic(err)
			}
			c.SendMsg(&Node{ipv4[0], msg.MyPort}, jsonMsg)
		default:
			err = c.callback(msg, &Node{ipv4[0], msg.MyPort})
			if err != nil {
				return err
			}
		}
	} else if rsp == OK_WITH_PAYLOAD {
		switch msg.Type {
		case CORE_LIST:
			log.Println("Refresh core node list ...")
			c.coreNodeSet.OverWriteByString(string(msg.Payload))
			log.Println("latest core node list:", c.coreNodeSet.GetNodesByString())
		default:
			err = c.callback(msg, nil)
			if err != nil {
				return err
			}
		}
	} else {
		panic(errors.Wrap(nil, "unknown response status"))
	}
	return nil
}

func (c *ConnectionManagerCore) addPeer(ip string, port uint16, msType MessageType) {
	portStr := strconv.Itoa(int(port))
	if msType == ADD {
		log.Printf("Adding Core: %s:%s", ip, portStr)
		c.coreNodeSet.Add(ip + ":" + portStr)
	} else if msType == ADD_AS_EDGE {
		log.Printf("Adding Edge: %s:%s", ip, portStr)
		c.edgeNodeSet.Add(ip + ":" + portStr)
	}
}

func (c *ConnectionManagerCore) removePeer(ip string, port uint16, msType MessageType) {
	portStr := strconv.Itoa(int(port))
	if msType == REMOVE {
		log.Printf("Removing Core: %s:%s", ip, portStr)
		c.coreNodeSet.Remove(ip + ":" + portStr)
	} else if msType == REMOVE_EDGE {
		c.edgeNodeSet.Remove(ip + ":" + portStr)
	}
}

func (c *ConnectionManagerCore) SendMsgToAllPeer(msg []byte) error {
	log.Println("Send message to all peer")
	nodes, err := c.coreNodeSet.GetNodes()
	if err != nil {
		return err
	}

	for _, peer := range nodes {
		if (*peer).IP != c.host || (*peer).Port != c.port {
			err = c.SendMsg(peer, msg)
			if err != nil {
				return err
				// TODO: continue with disconnection?
			}
		}
	}
	return nil
}
func (c *ConnectionManagerCore) SendMsgToAllEdge(msg []byte) error {
	log.Println("Send message to all edges")
	nodes, err := c.edgeNodeSet.GetNodes()
	if err != nil {
		return err
	}
	for _, edge := range nodes {
		err = c.SendMsg(edge, msg)
		if err != nil {
			return err
			// TODO: continue with disconnection?
		}
	}
	return nil
}

func (c *ConnectionManagerCore) SendMsg(peer *Node, msg []byte) error {
	conn, err := net.Dial("tcp", peer.String())
	if err != nil {
		return err
	}
	defer conn.Close()
	_, msgLog, _ := c.mm.Parse(msg)
	log.Println("send message:", msgLog)
	_, err = conn.Write(msg)
	return err
}

func (c *ConnectionManagerCore) checkPeersConnection() error {
	log.Println("check peer connections")
	currentCoreList, err := c.coreNodeSet.GetNodes()
	if err != nil {
		return err
	}
	outNodes := NewNodeSet()
	// TODO: can be multi thread
	for _, n := range currentCoreList {
		if n.Port == c.port && c.host == c.host {
			continue
		}
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
		jsonMsg, err := c.mm.Build(CORE_LIST, c.port, []byte(currentCoreListStr))
		if err != nil {
			panic(err)
		}
		c.SendMsgToAllEdge(jsonMsg)
		c.SendMsgToAllPeer(jsonMsg)
	}
	return nil
}

func (c *ConnectionManagerCore) checkEdgesConnection() error {
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

func (c *ConnectionManagerCore) ping(target *Node) error {
	conn, err := net.Dial("tcp", target.String())
	// TODO: error handling
	if err != nil {
		return err
	}
	defer conn.Close()
	// TODO: can be constant
	jsonMsg, _ := c.mm.Build(PING, c.port, nil)

	_, err = conn.Write(jsonMsg)
	if err != nil {
		return err
	}
	return nil
}
