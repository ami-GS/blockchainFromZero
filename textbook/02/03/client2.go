package main

import (
	"encoding/json"
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/02/03/core"
	"github.com/ami-GS/blockchainFromZero/textbook/02/03/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/02/03/p2p/message"
)

func main() {
	node, err := p2p.NodeFromString("192.168.1.51:50090")
	if err != nil {
		panic(err)
	}
	client := core.NewClientCore(50098, node)
	client.Start()

	data := map[string]string{
		"From":    "hoge",
		"To":      "fuga",
		"Message": "test",
	}

	enhancedMsg, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.ENHANCED, enhancedMsg)

	time.Sleep(5 * time.Second)
}
