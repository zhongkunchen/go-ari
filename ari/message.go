package ari

type Message struct {
	// doneChan is used to send SYN to message sender
	DoneChan    chan <-int

	serialNo    uint32

	MessageType []byte
	Body        []byte
	tags        [][]byte
	terms       map[[]byte][]byte
}

func NewMessage(doneChan chan <-int, serialNo uint32,
		messageType []byte, body []byte, tags[][]byte) *Message {
	m := &Message{
		DoneChan:doneChan,
		serialNo:serialNo,
		MessageType:messageType,
		Body:body,
		tags:tags,
	}
	return m
}

func (m *Message) SerialNo() uint32 {
	return m.serialNo
}

func (m *Message) AddTag(tag []byte) {
	m.tags = append(m.tags, tag)
}

func (m *Message) SetTerm(key []byte, value []byte)  {
	m.terms[key] = value
}

func (m *Message) GetTerm(key []byte) (value []byte, ok bool)  {
	value, ok = m.terms[key]
	return value, ok
}

// Done mark message finished
func (m *Message) Done()  {
	if m.DoneChan != nil {
		close(m.DoneChan)
		m.DoneChan = nil
	}
}
