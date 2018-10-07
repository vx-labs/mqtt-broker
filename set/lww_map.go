package set

import (
	"bytes"
	"encoding/gob"
	"errors"
	"sync"
)

type lwwMap struct {
	storage map[string]Entry
	mutex   sync.Mutex
}

var ErrNotFound = errors.New("not found")

func NewLWW() LWW {
	return &lwwMap{
		storage: map[string]Entry{},
	}
}

func DecodeLWW(data [][]byte) (LWW, error) {
	if len(data) < 1 {
		return nil, errors.New("invalid encoded payload size")
	}
	storage := map[string]Entry{}

	err := gob.NewDecoder(bytes.NewReader(data[0])).Decode(&storage)
	if err != nil {
		return nil, err
	}
	return &lwwMap{
		storage: storage,
	}, nil
}

func (lww *lwwMap) Serialize() [][]byte {
	lww.mutex.Lock()
	defer lww.mutex.Unlock()
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(lww.storage)
	if err == nil {
		return [][]byte{buf.Bytes()}
	}
	panic(err)
}

func (lww *lwwMap) get(id string) (Entry, error) {
	lww.mutex.Lock()
	defer lww.mutex.Unlock()
	entry, ok := lww.storage[id]
	if !ok {
		return Entry{}, ErrNotFound
	}
	return entry, nil
}

func (lww *lwwMap) insert(key string, entry Entry) error {
	lww.mutex.Lock()
	defer lww.mutex.Unlock()
	lww.storage[key] = entry
	return nil
}

func (lww *lwwMap) IsRemoved(value string) bool {
	entry, err := lww.get(value)
	if err == nil {
		return entry.IsDeleted()
	}
	return false
}
func (lww *lwwMap) IsSet(value string) bool {
	entry, err := lww.get(value)
	if err == nil {
		return entry.IsAdded()
	}
	return false
}

func (lww *lwwMap) Set(value string) error {
	return lww.insert(value, Entry{
		Add: now(),
	})
}

func (lww *lwwMap) Remove(value string) error {
	return lww.insert(value, Entry{
		Del: now(),
	})
}

func (lww *lwwMap) Iterate(f func(string, Entry) error) error {
	copy := lww.Copy()
	for key, entry := range copy.(*lwwMap).storage {
		err := f(key, entry)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (s *lwwMap) Copy() LWW {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	copy := &lwwMap{
		storage: make(map[string]Entry, len(s.storage)),
	}
	for key, value := range s.storage {
		copy.storage[key] = value
	}
	return copy
}
func (s *lwwMap) Length() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return len(s.storage)
}
func (s *lwwMap) Merge(r LWW) LWW {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delta := &lwwMap{
		storage: map[string]Entry{},
	}
	r.Iterate(func(key string, rt Entry) error {
		t, _ := s.storage[key]

		if t.Add < rt.Add {
			t.Add = rt.Add
		} else {
			rt.Add = 0
		}

		if t.Del < rt.Del {
			t.Del = rt.Del
		} else {
			rt.Del = 0
		}

		if !rt.IsZero() {
			s.storage[key] = t
			delta.storage[key] = t
		}
		return nil
	})
	return delta
}
