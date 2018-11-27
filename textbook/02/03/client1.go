package main

import (
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/02/03/core"
	"github.com/ami-GS/blockchainFromZero/textbook/02/03/p2p"
)

func main() {
	node, err := p2p.NodeFromString("192.168.1.51:50082")
	if err != nil {
		panic(err)
	}
	client := core.NewClientCore(50100, node)
	client.Start()

	time.Sleep(5 * time.Second)
}
