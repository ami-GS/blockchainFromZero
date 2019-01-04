package main

import (
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/03/04/core"
	"github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p"
)

func main() {
	node, err := p2p.NodeFromString("192.168.1.8:50051")
	//node, err := p2p.NodeFromString("10.32.45.12:50082")
	if err != nil {
		panic(err)
	}
	server := core.NewServerCore(50053, node)
	_, cancel := server.Start()
	defer cancel()
	err = server.JoinNetwork()
	if err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Second)
	// method name will be changed
	//server.GetAllChainsForResolveConflict()

	c := make(<-chan struct{}, 0)
	<-c
}
