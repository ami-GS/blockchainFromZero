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
	client := core.NewClientCore(50095, node)
	client.Start()

	time.Sleep(10 * time.Second)

	txMap := map[string]string{
		"Sender":    "test4",
		"Recipient": "test5",
		"Value":     "3",
	}

	enhancedMsg, err := json.Marshal(txMap)
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)

	txMap2 := map[string]string{
		"Sender":    "test6",
		"Recipient": "test7",
		"Value":     "2",
	}

	enhancedMsg, err = json.Marshal(txMap2)
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)

	time.Sleep(10 * time.Second)

	txMap3 := map[string]string{
		"Sender":    "test8",
		"Recipient": "test9",
		"Value":     "10",
	}

	enhancedMsg, err = json.Marshal(txMap3)
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)

	time.Sleep(5 * time.Second)
}
