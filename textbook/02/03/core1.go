package main

import (
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/02/03/core"
)

func main() {
	server := core.NewServerCore(50082, nil)
	server.Start()
	time.Sleep(10 * time.Second)
}
