package filebeater

import (
	"bytes"
	"regexp"
	"sort"
)

// Codec is used to scan logs from byte slice
type Codec interface {
	NextLogs([]byte) (logs [][]byte)
}

func copyFrom(bs []byte) []byte{
	if len(bs) == 0 {
		return nil
	}
	body := make([]byte, len(bs))
	copy(body, bs)
	return body
}

var _ Codec = &MultiLineCodec{}

// MultiLineCodec is a codec to parse multi lines byte stream
type MultiLineCodec struct {
	tokenReg *regexp.Regexp
	buf *bytes.Buffer
}

// NewMultiLineCodec creates a new `MultiLineCodec` instance
// with the every regexp `pattern`
func NewMultiLineCodec(pattern string) (*MultiLineCodec, error) {
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	c := &MultiLineCodec{
		tokenReg: reg,
		buf: bytes.NewBuffer(make([]byte, 1024 * 64)),
	}
	c.buf.Reset()
	return c, nil
}

// Done wraps pending bytes as a log []byte
func (c *MultiLineCodec) Done() []byte{
	defer c.buf.Reset()
	if c.buf.Len() != 0 {
		return copyFrom(c.buf.Bytes())
	}
	return nil
}

// NextLogs scans logs bytes from codec buf and the arg bs
func (c *MultiLineCodec) NextLogs(bs []byte) (logs [][]byte) {
	if bs == nil {
		return nil
	}
	matches := c.tokenReg.FindAllIndex(bs, -1)
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
	logs = make([][]byte, len(indices))
	var start, i, n int = 0, 0, 0

	// the bs maybe format below
	// 1. xxixxxixxxxix...
	// 2. ixxxixxxxxiixi
	// 3. ixxxixxxxxiix...
	for n, i = range indices {
		if i != 0 {
			c.buf.Write(bs[start:i])
		}
		logs[n] = copyFrom(c.buf.Bytes())
		c.buf.Reset()
		start = i
	}
	c.buf.Write(bs[start:])
	// maybe the first msg is nil
	if logs[0] == nil {
		logs = logs[1:]
	}
	if len(logs) == 0 {
		logs = nil
	}
	return logs
}

