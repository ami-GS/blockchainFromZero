package p2p

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"

	. "github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p/message"
	"github.com/pkg/errors"
)

type ConnectionManager4Edge struct {
	*ConnectionManager
}

func NewConnectionManager4Edge(host string, port uint16, bootstrap *Node, apiCallback func(*Message, *Node) error) ConnectionManagerI {
	return &ConnectionManager4Edge{
		NewConnectionManager(host, port, bootstrap, apiCallback, false),
	}
}

func (c *ConnectionManager4Edge) Start(ctx context.Context) {
	go c.waitForAccess(ctx)
	go func() {
		pingToCore := time.NewTicker(PING_INTERVAL * time.Second)
		for {
			select {
			case <-pingToCore.C:
				err := c.ping()
				if err != nil {
					panic(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *ConnectionManager4Edge) waitForAccess(ctx context.Context) {
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
			log.Printf("Connected by ... %v\n", conn.RemoteAddr())
			go c.handleMessage(conn)
		}
	}
}

func (c *ConnectionManager4Edge) ConnectionClose() error {
	// TODO: send cancel to waitForAccess
	// TODO: send cancel to pingToCore.C of Start()
	return nil
}

func (c *ConnectionManager4Edge) handleMessage(conn net.Conn) error {
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
		if msg.Type != PING {
			panic(errors.Wrap(nil, fmt.Sprintf("Edge does not have handlers of this message: %d", msg.Type)))
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

func (c *ConnectionManager4Edge) SendMsg(peer *Node, msg []byte) error {
	fallback := func() {
		log.Println("Connection failed for peer:", &peer)
		c.coreNodeSet.Remove(peer.String())
		log.Println("Tring to connecto into P2P netowork ...")
		nodes, err := c.coreNodeSet.GetNodes()
		if err != nil {
			panic(err)
		}
		if len(nodes) > 0 {
			c.JoinNetwork(nodes[0])
			c.SendMsg(peer, msg)
		} else {
			log.Println("No core node found in our list ...")
			// TODO: cancel ping
		}
	}

	conn, err := net.Dial("tcp", peer.String())
	if err != nil {
		fallback()
		return err
	}
	defer conn.Close()
	_, msgLog, _ := c.mm.Parse(msg)
	log.Println("send message:", msgLog)
	_, err = conn.Write(msg)
	if err != nil {
		fallback()
		return err
	}
	return nil
}

func (c *ConnectionManager4Edge) ping() error {
	target := c.bootstrapNode
	fallback := func() error {
		targetStr := target.String()
		log.Printf("Connection failed to: %s\n", targetStr)
		c.coreNodeSet.Remove(targetStr)
		log.Println("Tring to connect into P2P network ...")
		currentCoreList, err := c.coreNodeSet.GetNodes()
		if err != nil {
			panic(err)
		}
		if len(currentCoreList) == 0 {
			log.Println("No Core node found anymore")
			return errors.Wrap(nil, "No Core node found anymore in this edge")
		}
		c.bootstrapNode = currentCoreList[0]
		err = c.connectToP2PNW(currentCoreList[0])
		return err
	}
	conn, err := net.Dial("tcp", target.String())
	// TODO: error handling
	if err != nil {
		return fallback()
	}
	defer conn.Close()
	// TODO: can be constant
	jsonMsg, _ := c.mm.Build(PING, c.port, nil)

	_, err = conn.Write(jsonMsg)
	if err != nil {
		return fallback()
	}
	return nil
}
