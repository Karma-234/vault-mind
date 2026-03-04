package storage

import (
	"time"

	"github.com/cockroachdb/pebble"
)

type Credential struct {
	ID      string    `json:"id"`
	Service string    `json:"service"`
	Type    string    `json:"type"`
	Notes   string    `json:"notes,omitempty"`
	Created time.Time `json:"createdAt"`
	Updated time.Time `json:"updatedAt,omitempty"`
}
type Storage interface {
	AddCredential(service, credType, secret, notes string) (string, error)
	ListCredentials() ([]Credential, error)
	GetCredential(id string) (*Credential, string, error)
}

type VaultPebbleStorage struct {
	Storage
	db  *pebble.DB
	key []byte
}

func (s *VaultPebbleStorage) AddCredential(service, credType, secret, notes string) (string, error) {

	return "", nil
}
