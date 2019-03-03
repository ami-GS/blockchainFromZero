package message

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
	NEW_TRANSACTION
	NEW_BLOCK
	REQUEST_FULL_CHAIN
	RSP_FULL_CHAIN
	ENHANCED // includes cipher message
)

func (m MessageType) String() string {
	return []string{
		"ADD",
		"REMOVE",
		"CORE_LIST",
		"REQUEST_CORE_LIST",
		"PING",
		"ADD_AS_EDGE",
		"REMOVE_EDGE",
		"NEW_TRANSACTION",
		"NEW_BLOCK",
		"REQUEST_FULL_CHAIN",
		"RSP_FULL_CHAIN",
		"ENHANCED", // includes cipher message
	}[int(m)]
}

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
	MyPort   uint16
	// TODO: need to be interface? any type comming
	Payload []byte
}

func (m *Message) String() string {
	var payload string
	if len(m.Payload) >= 50 {
		payload = string(m.Payload[:45]) + " ..."
	} else {
		payload = string(m.Payload)
	}
	return fmt.Sprintf("\n\tProtocol: %s\n\tVersion: %s\n\tType: %s\n\tPortFrom: %d\n\tPayload: %s", m.Protocol, m.Version, m.Type, m.MyPort, payload)
}

func build(msgType MessageType, myPort uint16, payload []byte) ([]byte, error) {
	msg := Message{
		Protocol: PROTOCOL_NAME,
		Version:  VERSION,
		Type:     msgType,
		MyPort:   myPort,
		Payload:  payload,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return jsonMsg, err
}

func parse(msg []byte) (MessageResponse, *Message, error) {
	var outMsg Message
	err := json.Unmarshal(msg, &outMsg)
	if err != nil {
		panic(err)
	}

	if outMsg.Protocol != PROTOCOL_NAME {
		return NONE, nil, errors.Wrap(nil, "Protocol mismatch")
	} else if outMsg.Version != VERSION {
		return NONE, nil, errors.Wrap(nil, "Protocol version mismatch")
	}
	// TODO: check if payload is really proviced
	switch outMsg.Type {
	case CORE_LIST, NEW_TRANSACTION, NEW_BLOCK, RSP_FULL_CHAIN, ENHANCED, ADD_AS_EDGE:
		return OK_WITH_PAYLOAD, &outMsg, nil
	default:
		return OK_WITHOUT_PAYLOAD, &outMsg, nil
	}
}
