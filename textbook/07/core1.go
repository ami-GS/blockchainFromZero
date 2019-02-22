package main

import (
	"github.com/ami-GS/blockchainFromZero/textbook/07/core"
)

func main() {
	server := core.NewServerCore(50051, nil)
	_, cancel := server.Start()
	defer cancel()
	c := make(<-chan struct{}, 0)
	<-c
}
