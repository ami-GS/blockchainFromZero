package main

import (
	"encoding/json"
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/03/03/core"
	"github.com/ami-GS/blockchainFromZero/textbook/03/03/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/03/03/p2p/message"
)

func main() {
	node, err := p2p.NodeFromString("192.168.1.8:50082")
	if err != nil {
		panic(err)
	}
	client := core.NewClientCore(50088, node)
	client.Start()

	txMap := map[string]string{
		"Sender":    "test1",
		"Recipient": "test2",
		"Value":     "3",
	}

	enhancedMsg, err := json.Marshal(txMap)
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)

	txMap2 := map[string]string{
		"Sender":    "test1",
		"Recipient": "test3",
		"Value":     "2",
	}

	enhancedMsg, err = json.Marshal(txMap2)
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)
	txMap3 := map[string]string{
		"Sender":    "test5",
		"Recipient": "test6",
		"Value":     "10",
	}

	enhancedMsg, err = json.Marshal(txMap3)
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)

	time.Sleep(5 * time.Second)
}
