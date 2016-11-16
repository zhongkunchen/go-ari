package ari

type Sender interface {
	Send(msg *Message)
}
