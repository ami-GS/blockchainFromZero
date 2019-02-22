package p2p

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	. "github.com/ami-GS/blockchainFromZero/textbook/06/p2p/message"
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
	return c.sendMsg(c.bootstrapNode, jsonMsg)
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
	ipv4 := strings.Split(conn.RemoteAddr().String(), ":")
	if rsp == OK_WITHOUT_PAYLOAD {
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

			if msg.Type == ADD {
				log.Println("ADD node request")
				return c.SendMsgBroadcast([]byte(nodesStr), CORE_LIST)
			}
			log.Println("ADD_AS_EDGE request")
			return c.SendMsgTo([]byte(nodesStr), CORE_LIST, &Node{ipv4[0], msg.MyPort})
		case REMOVE:
			log.Printf("REMOVE request from %d\n", ipv4[0], msg.MyPort)
			c.removePeer(ipv4[0], msg.MyPort, msg.Type)
			nodesStr := c.coreNodeSet.GetNodesByString()
			return c.SendMsgBroadcast([]byte(nodesStr), CORE_LIST)
		case REMOVE_EDGE:
			log.Printf("REMOVE_EDGE request from %d\n", ipv4[0], msg.MyPort)
			c.removePeer(ipv4[0], msg.MyPort, msg.Type)
		case PING:
		case REQUEST_CORE_LIST:
			log.Println("List for Core nodes was requested")
			nodesStr := c.coreNodeSet.GetNodesByString()
			return c.SendMsgTo([]byte(nodesStr), CORE_LIST, &Node{ipv4[0], msg.MyPort})
		default:
			err = c.callback(msg, &Node{ipv4[0], msg.MyPort})
			if err != nil {
				return err
			}
		}
	} else if rsp == OK_WITH_PAYLOAD {
		switch msg.Type {
		case CORE_LIST:
			if c.coreNodeSet.Len() == 1 {
				// from bootstrap node at beginning
				if c.bootstrapNode.IP != ipv4[0] || c.bootstrapNode.Port != msg.MyPort {
					log.Println("received unsafe core node list from %s%d", ipv4[0], msg.MyPort)
					return nil
				}
			} else {
				if isCore := c.coreNodeSet.Has(Node{ipv4[0], msg.MyPort}); !isCore {
					log.Println("CORE_LIST from unknown node: %s:%d", ipv4[0], msg.MyPort)
					return nil
				}
				// TODO: inefficient
				nodes := strings.Split(string(msg.Payload), ",")
				for _, nodeStr := range nodes {
					out := strings.Split(nodeStr, ":")
					port, err := strconv.Atoi(out[1])
					if err != nil {
						panic(err)
					}
					if out[0] == c.host && uint16(port) == c.port {
						continue
					}
					node := &Node{out[0], uint16(port)}
					err = c.ping(node)
					if err != nil {
						log.Println("received unsafe core node list from", ipv4[0], msg.MyPort)
						return nil
					}
				}
			}
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

func (c *ConnectionManagerCore) SendMsgTo(data []byte, mType MessageType, to *Node) error {
	outMsg, err := c.GetMessageBytes(mType, data)
	if err != nil {
		return err
	}
	return c.sendMsg(to, outMsg)
}

func (c *ConnectionManagerCore) sendMsgToAllPeer(msg []byte) error {
	nodes, err := c.coreNodeSet.GetNodes()
	if err != nil {
		return err
	}
	for _, peer := range nodes {
		if (*peer).IP != c.host || (*peer).Port != c.port {
			err = c.sendMsg(peer, msg)
			if err != nil {
				return err
				// TODO: continue with disconnection?
			}
		}
	}
	return nil
}
func (c *ConnectionManagerCore) SendMsgToAllPeer(data []byte, mType MessageType) error {
	log.Println("Send message to all peer")
	msg, err := c.GetMessageBytes(mType, data)
	if err != nil {
		return err
	}
	return c.sendMsgToAllPeer(msg)
}

func (c *ConnectionManagerCore) sendMsgToAllEdge(msg []byte) error {
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
func (c *ConnectionManagerCore) SendMsgToAllEdge(data []byte, mType MessageType) error {
	log.Println("Send message to all edges")
	msg, err := c.GetMessageBytes(mType, data)
	if err != nil {
		return err
	}
	return c.sendMsgToAllEdge(msg)
}

func (c *ConnectionManagerCore) SendMsgBroadcast(data []byte, mType MessageType) error {
	log.Println("Send message to all edges")
	msg, err := c.GetMessageBytes(mType, data)
	if err != nil {
		return err
	}
	err = c.sendMsgToAllPeer(msg)
	if err != nil {
		return err
	}
	return c.sendMsgToAllEdge(msg)
}

func (c *ConnectionManagerCore) sendMsg(peer *Node, msg []byte) error {
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
		return c.SendMsgBroadcast([]byte(currentCoreListStr), CORE_LIST)
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

	//return c.SendMsgTo(nil, PING, target)
	return nil
}
