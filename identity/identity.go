package identity

import (
	"crypto/sha1"
	"fmt"
)

// Identity represents a Host network identity, with a unique ID, and two network identities: one private, and one public.
type Identity interface {
	ID() string
	WithID(string) Identity
	Private() Address
	Public() Address
}

type identity struct {
	id      string
	private *address
	public  *address
}

func (i identity) WithID(id string) Identity {
	i.id = id
	return &i
}
func (i identity) ID() string {
	if i.id != "" {
		return i.id
	}
	hash := sha1.New()
	hash.Write([]byte(i.public.String()))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
func (i identity) Private() Address {
	return i.private
}
func (i identity) Public() Address {
	return i.public
}
