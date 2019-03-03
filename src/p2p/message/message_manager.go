package message

type MessageManager struct{}

// TODO: MessageManager is not needed?
func (m *MessageManager) Build(msgType MessageType, myPort uint16, payload []byte) ([]byte, error) {
	return build(msgType, myPort, payload)
}

func (m *MessageManager) Parse(msg []byte) (MessageResponse, *Message, error) {
	return parse(msg)
}
