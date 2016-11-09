package filebeater

import (
	"github.com/argpass/go-ari/ari"
	"bytes"
	"regexp"
	"sort"
)

type Codec interface {
	NextMessages([]byte) (msgs []*ari.Message)
}

// MessageFrom wraps a `Message` from []byte
func MessageFrom(bs []byte) *ari.Message  {
	if len(bs) == 0 {
		return nil
	}
	body := make([]byte, len(bs))
	copy(body, bs)
	msg := &ari.Message{Body:body}
	return msg
}

var _ Codec = &MultiLineCodec{}

type MultiLineCodec struct {
	tokenReg *regexp.Regexp
	buf *bytes.Buffer
}

func NewMultiLineCodec(pattern string) (*MultiLineCodec, error) {
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	c := &MultiLineCodec{
		tokenReg: reg,
		buf: bytes.NewBuffer(make([]byte, 1024 * 64)),
	}
	return c
}

func (c *MultiLineCodec) Done() *ari.Message {
	defer c.buf.Reset()
	if c.buf.Len() != 0 {
		return MessageFrom(c.buf.Bytes())
	}
	return nil
}

func (c *MultiLineCodec) NextMessages(bs []byte) (msgs []*ari.Message) {
	if bs == nil {
		return nil
	}
	matches := c.tokenReg.FindAllIndex(bs, 0)
	if matches == nil {
		c.buf.Write(bs)
		return nil
	}
	if len(matches) == 0 {
		return nil
	}
	indices := make([]int, len(matches))
	for i, match := range matches {
		indices[i] = match[0]
	}
	// sort indices in increasing order
	sort.Ints(indices)
	msgs = make([]*ari.Message, len(indices))
	var start, i, n int = 0, 0, 0

	// the bs maybe format below
	// 1. xxixxxixxxxix...
	// 2. ixxxixxxxxiixi
	// 3. ixxxixxxxxiix...
	for n, i = range indices {
		if i != 0 {
			c.buf.Write(bs[start:i])
		}
		msgs[n] = MessageFrom(c.buf.Bytes())
		c.buf.Reset()
		start = i
	}
	c.buf.Write(bs[start:])
	// maybe the first msg is nil
	if msgs[0] == nil {
		msgs = msgs[1:]
	}
	if len(msgs) == 0 {
		msgs = nil
	}
	return msgs
}
