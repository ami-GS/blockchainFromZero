package p2p

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type MessageType uint16

const (
	ADD MessageType = iota
	REMOVE
	CORE_LIST
	REQUEST_CORE_LIST
	PING
	ADD_AS_EDGE
	REMOVE_EDGE
)

type MessageError uint16

const (
	PROTOCOL_MISMATCH MessageError = iota
	VERSION_MISMATCH
)

type MessageResponse uint16

const (
	OK_WITH_PAYLOAD MessageResponse = iota
	OK_WITHOUT_PAYLOAD
	NONE
)

const PROTOCOL_NAME = "simple_bitcoin_protocol"
const VERSION = "0.1.0"

type Message struct {
	Protocol string
	Version  string
	Type     MessageType
	myPort   uint16
	payload  string
}

type MessageManager struct{}

func (m *MessageManager) Build(msgType MessageType, myPort uint16, payload string) ([]byte, error) {
	msg := Message{
		Protocol: PROTOCOL_NAME,
		Version:  VERSION,
		Type:     msgType,
		myPort:   myPort,
		payload:  payload,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonMsg))
	return jsonMsg, err
}

func (m *MessageManager) Parse(msg []byte) (MessageResponse, *Message, error) {
	var outMsg Message
	fmt.Println(string(msg))
	err := json.Unmarshal(msg, &outMsg)
	if err != nil {
		panic(err)
	}

	if outMsg.Protocol != PROTOCOL_NAME {
		return NONE, nil, errors.Wrap(nil, "Protocol mismatch")
	} else if outMsg.Version != VERSION {
		return NONE, nil, errors.Wrap(nil, "Protocol version mismatch")
	}
	switch outMsg.Type {
	case CORE_LIST:
		return OK_WITH_PAYLOAD, &outMsg, nil
	default:

	}
	return OK_WITHOUT_PAYLOAD, &outMsg, nil
}
