package subscriptions

import "bytes"

var (
	SEP = byte('/')
)

type Topic []byte

func (t Topic) Chop() (Topic, []byte, bool) {
	end := bytes.IndexByte(t, SEP)
	if end < 0 {
		return nil, t, false
	}
	return t[end+1:], t[:end], true
}
