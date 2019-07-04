package transport

import (
	"crypto/tls"
	"io"
	"time"
)

type TimeoutReadWriteCloser interface {
	SetDeadline(time.Time) error
	io.ReadWriteCloser
}

type Metadata struct {
	Name            string
	Encrypted       bool
	EncryptionState *tls.ConnectionState
	RemoteAddress   string
	Channel         TimeoutReadWriteCloser
	Endpoint        string
}
