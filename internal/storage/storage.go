package storage

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/karma-234/vault-mind-mcp/internal/crypto"
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
	mu  sync.RWMutex
}

func NewVaultPebbleStorage(dbPath string, key []byte) (*VaultPebbleStorage, error) {

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}
	db, err := pebble.Open(dbPath, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &VaultPebbleStorage{
		db:  db,
		key: key,
	}, nil
}
func generateID() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 16
	id := make([]byte, length)
	for i := range id {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		id[i] = charset[num.Int64()]
	}
	return string(id), nil
}
func (s *VaultPebbleStorage) AddCredential(service, credType, secret, notes string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	encSecret, err := crypto.Encrypt(s.key, []byte(secret))
	if err != nil {
		return "", err
	}
	id, err := generateID()
	if err != nil {
		return "", err
	}
	cred := Credential{
		ID:      id,
		Service: service,
		Type:    credType,
		Notes:   notes,
		Created: time.Now(),
	}
	data, err := json.Marshal(cred)
	if err != nil {
		return "", err
	}
	batch := s.db.NewBatch()
	defer batch.Close()
	if err := batch.Set([]byte("Meta:"+id), data, pebble.Sync); err != nil {
		return "", err
	}
	if err := batch.Set([]byte("Secret:"+id), encSecret, pebble.Sync); err != nil {
		return "", err
	}
	if err := batch.Commit(pebble.Sync); err != nil {
		return "", err
	}
	return id, nil
}
