package main

import (
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/03/04/core"
	"github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/03/04/p2p/message"
	"github.com/ami-GS/blockchainFromZero/textbook/03/04/transaction"
)

func main() {
	node, err := p2p.NodeFromString("192.168.1.8:50053")
	//node, err := p2p.NodeFromString("10.32.45.12:50082")
	if err != nil {
		panic(err)
	}
	client := core.NewClientCore(50093, node)
	_, cancel := client.Start()
	defer cancel()

	time.Sleep(2 * time.Second)

	txMap := transaction.New("hoge3", "hoge31", "33")
	enhancedMsg, err := txMap.GetJson()
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)

	time.Sleep(2 * time.Second)

	txMap2 := transaction.New("hoge3", "hoge32", "333")

	enhancedMsg, err = txMap2.GetJson()
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)

	time.Sleep(3 * time.Second)

	txMap3 := transaction.New("hoge3", "hoge33", "3333")
	enhancedMsg, err = txMap3.GetJson()
	if err != nil {
		panic(err)
	}
	client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)

	loop := 0
	for {

		time.Sleep(20 * time.Second)
		txMap4 := transaction.New("hoge3", "hoge33", 3333+loop)
		enhancedMsg, err = txMap4.GetJson()
		if err != nil {
			panic(err)
		}
		client.SendMessageToMyCore(message.NEW_TRANSACTION, enhancedMsg)
		loop++
	}

}
