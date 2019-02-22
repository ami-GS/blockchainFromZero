package message

import (
	"bytes"
	"sync"
)

// Storig Message, but comparing only payload for now
// TODO: could be optimized by hashmap, or from/to pubkey
// TODO: should have timeout for each message
type MessageList struct {
	mu   *sync.Mutex
	list []Message
}

func (l *MessageList) Add(msg Message) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// TODO: start timer for each msg to remove automatically
	l.list = append(l.list, msg)
}

func (l *MessageList) Has(msg *Message) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, m := range l.list {
		if bytes.Equal(msg.Payload, m.Payload) {
			return true
		}
	}
	return false
}

func (l *MessageList) Index(msg *Message) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, m := range l.list {
		if bytes.Equal(msg.Payload, m.Payload) {
			return i
		}
	}
	return -1
}

func (l *MessageList) Remove(msg Message) {
	l.mu.Lock()
	defer l.mu.Unlock()
	index := l.Index(&msg)
	l.list = append(l.list[:index], l.list[index+1:]...)
}
