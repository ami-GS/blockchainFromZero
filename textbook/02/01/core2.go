package main

import (
	"github.com/ami-GS/blockchainFromZero/textbook/02/01/core"
	"github.com/ami-GS/blockchainFromZero/textbook/02/01/p2p"
)

func main() {
	node, err := p2p.NodeFromString("192.168.1.51:50082")
	if err != nil {
		panic(err)
	}
	server := core.NewServerCore(50090, node)
	server.Start()
	err = server.JoinNetwork()
	if err != nil {
		panic(err)
	}
}
