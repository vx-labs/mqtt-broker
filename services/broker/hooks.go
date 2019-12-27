package broker

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
)

func newUUID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("failed to read random bytes: %v", err))
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}

func makeSubID(session string, pattern []byte) string {
	hash := sha1.New()
	_, err := hash.Write([]byte(session))
	if err != nil {
		return ""
	}
	_, err = hash.Write(pattern)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getLowerQoS(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
