package main

import (
	"github.com/ami-GS/blockchainFromZero/textbook/06/core"
	"github.com/ami-GS/blockchainFromZero/textbook/06/p2p"
)

func main() {
	node, err := p2p.NodeFromString("192.168.1.12:50051")
	//node, err := p2p.NodeFromString("10.32.45.12:50082")
	if err != nil {
		panic(err)
	}
	server := core.NewServerCore(50052, node)
	_, cancel := server.Start()
	defer cancel()
	err = server.JoinNetwork()
	if err != nil {
		panic(err)
	}
	c := make(<-chan struct{}, 0)
	<-c
}
