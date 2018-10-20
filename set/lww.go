package set

import "time"

type LWW interface {
	RemoveOlder(maxage int64) LWW
	Remove(value string) error
	Set(value string) error
	IsSet(value string) bool
	IsRemoved(value string) bool
	Iterate(func(string, Entry) error) error
	Merge(LWW) LWW
	Copy() LWW
	Serialize() [][]byte
	Length() int
}

type Entry struct {
	Add int64
	Del int64
}

func (e Entry) IsZero() bool {
	return e.Add == 0 && e.Del == 0
}
func (e Entry) IsAdded() bool {
	return e.Add > 0 && e.Add > e.Del
}
func (e Entry) IsDeleted() bool {
	return e.Del > 0 && e.Del > e.Add
}

func now() int64 {
	return time.Now().UnixNano()
}
