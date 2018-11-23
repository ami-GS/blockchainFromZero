package p2p

import (
	"log"
	"strconv"
	"strings"
	"sync"
)

type Node struct {
	IP   string
	port uint16
}

func NodeFromString(nodeString string) (*Node, error) {
	out := strings.Split(nodeString, ":")
	portInt, err := strconv.Atoi(out[1])
	if err != nil {
		return nil, err
	}
	return &Node{
		IP:   out[0],
		port: uint16(portInt),
	}, nil
}

func (n Node) String() string {
	return n.IP + ":" + strconv.Itoa(int(n.port))
}

type NodeSet struct {
	mu   *sync.Mutex
	list map[string]struct{} // set
}

// TODO: tobe node.NewSet?
func NewNodeSet() *NodeSet {
	return &NodeSet{
		mu:   new(sync.Mutex),
		list: make(map[string]struct{}),
	}
}

func (c *NodeSet) Add(network string) {
	log.Println("Adding peer:", network)
	c.mu.Lock()
	c.list[network] = struct{}{}
	c.mu.Unlock()
	log.Println("Current list:", c.list)
}

func (c *NodeSet) Remove(network string) {
	c.mu.Lock()
	delete(c.list, network)
	c.mu.Unlock()
}

func (c *NodeSet) GetNodesByString() string {
	out := ""
	for k, _ := range c.list {
		out += k + ","
	}
	return out[:len(out)-1]
}

func (c *NodeSet) OverWriteByString(nodes string) {
	nodeSlice := strings.Split(nodes, ",")
	c.mu.Lock()
	c.list = make(map[string]struct{})
	for _, node := range nodeSlice {
		c.list[node] = struct{}{}
	}
	c.mu.Unlock()
}

func (c *NodeSet) GetNodes() ([]*Node, error) {
	out := make([]*Node, len(c.list))
	c.mu.Lock()
	defer c.mu.Unlock()
	i := 0
	var err error
	for k := range c.list {
		out[i], err = NodeFromString(k)
		if err != nil {
			// TODO: these errors should be pack as slice?
			panic(err)
		}
		i++
	}
	return out, nil
}

type CoreNodeSet struct {
	*NodeSet
}
type EdgeNodeSet struct {
	*NodeSet
}
