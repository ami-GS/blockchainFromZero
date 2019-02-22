package message

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/pkg/errors"
)

type MessageHandler struct {
	msgList *MessageList
}

type RequestType uint8

const (
	SEND_TO RequestType = iota
	SEND_TO_ALL_PEER
	SEND_TO_ALL_EDGE
	PASS_TO_CLIENT_API
	GET_API_ORIGIN
)

type APIResponseType uint8

const (
	API_OK APIResponseType = iota
	API_ERROR
	SERVER_CORE_API
	CLIENT_CORE_API
)

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		msgList: &MessageList{
			mu:   new(sync.Mutex),
			list: make([]Message, 0),
		},
	}
}

func (h *MessageHandler) HandleMessage(msg *Message, apiCallback func(request RequestType, msg *Message) (APIResponseType, error)) error {
	if h.msgList.Has(msg) {
		log.Println("Drop message, duplication")
		// avoid duplication
		return nil
	}
	h.msgList.Add(*msg)

	resp, err := apiCallback(GET_API_ORIGIN, nil)
	if err != nil {
		return err
	}
	switch resp {
	case SERVER_CORE_API:
		// TODO: inefficient, unmarshaling in API() as well
		enhancedMsg := make(map[string]string)
		json.Unmarshal(msg.Payload, &enhancedMsg)
		if mType, ok := enhancedMsg["MessageType"]; ok && mType == "Cipher" {
			log.Println("Received Cipher Message")
			status, _ := apiCallback(SEND_TO, msg)
			if status == API_ERROR {
				log.Println("Bloadcast from Core:", *msg)
				_, err := apiCallback(SEND_TO_ALL_PEER, msg)
				if err != nil {
					return err
				}
			}
		} else {
			log.Println("Bloadcast from Core:", *msg)
			_, err := apiCallback(SEND_TO_ALL_PEER, msg)
			if err != nil {
				return err
			}
			_, err = apiCallback(SEND_TO_ALL_EDGE, msg)
			return err
		}
	case CLIENT_CORE_API:
		log.Println("Client custom functionality called:", *msg)
		_, err = apiCallback(PASS_TO_CLIENT_API, msg)
		return err
	case API_OK, API_ERROR:
		// pass
	default:
		return errors.Wrap(nil, "Unknown API response")
	}
	return nil
}
