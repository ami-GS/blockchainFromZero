package main

import (
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/03/03/core"
	"github.com/ami-GS/blockchainFromZero/textbook/03/03/p2p"
)

func main() {
	node, err := p2p.NodeFromString("192.168.1.8:50082")
	if err != nil {
		panic(err)
	}
	server := core.NewServerCore(50090, node)
	server.Start()
	err = server.JoinNetwork()
	if err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second)
}
