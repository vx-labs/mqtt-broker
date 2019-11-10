package pb

import (
	"bytes"
	"io"
)

var (
	SEP = byte('/')
)

func NewTopic(topic []byte) *Topic {
	return &Topic{
		payload: topic,
		len:     len(topic),
	}
}

type Topic struct {
	payload []byte
	len     int
	start   int // current token start separator
	end     int // current token end separator
}

func (t *Topic) Bytes() []byte {
	return t.payload
}

// Next advances to the next token, and return the read token
func (t *Topic) Next() error {
	if t.end == t.len {
		return io.EOF
	}
	t.start = t.end
	cursor := t.start
	end := bytes.IndexByte(t.payload[cursor:], SEP) + cursor + 1
	if end == cursor {
		t.end = t.len
		return nil
	}
	t.end = end
	return nil
}

// Head returns the previous tokens and separators
func (t *Topic) Head() []byte {
	return t.payload[:t.start]
}
func (t *Topic) Cur() []byte {
	if t.end == t.len {
		return t.payload[t.start:]
	}
	return t.payload[t.start : t.end-1]
}

// Head returns the tokens and separators after the cursor
func (t *Topic) Tail() []byte {
	return t.payload[t.end:]
}

func (t *Topic) Chop() (*Topic, string, bool) {
	for idx, char := range t.payload {
		if char == SEP {
			token := string(t.payload[0:idx])
			return NewTopic(t.payload[idx+1:]), token, true
		}
	}
	return nil, string(t.payload), false
}
func (t *Topic) Length() int {
	count := 1
	for _, c := range t.payload {
		if c == SEP {
			count++
		}
	}
	return count
}

func (t *Topic) Match(pattern []byte) bool {
	switch bytes.Compare(t.payload, pattern) {
	case 0:
		return true
	case -1:
		return false
	}
	target := NewTopic(pattern)
	self := NewTopic(t.payload)
	var err error
	for {
		err = target.Next()
		if err != self.Next() {
			return false
		}
		if err != nil {
			return true
		}
		t := target.Cur()
		switch t[0] {
		case '+':
		case '#':
			return true
		default:
			s := self.Cur()
			if bytes.Compare(t, s) != 0 {
				return false
			}
		}
	}
}
