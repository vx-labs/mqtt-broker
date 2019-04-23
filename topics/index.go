package topics

import memdb "github.com/hashicorp/go-memdb"

type ByteSliceIndexer struct {
	i memdb.StringFieldIndex
}

func (b *ByteSliceIndexer) FromArgs(opts ...interface{}) ([]byte, error) {
	return b.i.FromArgs(opts...)
}

func (b *ByteSliceIndexer) FromObject(obj interface{}) (bool, []byte, error) {
	message := obj.(*Metadata)
	return true, append(message.GetTopic(), '\x00'), nil
}
