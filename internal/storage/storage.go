package storage

import (
	"crypto/rand"
	"encoding/json"
	"log"
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
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, dbPath)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}

	db, err := pebble.Open(path, &pebble.Options{})
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

func (s *VaultPebbleStorage) AddBulkCredentials(creds []Credential) error {
	s.mu.Lock()
	batch := s.db.NewBatch()
	defer batch.Close()
	defer s.mu.Unlock()
	for _, cred := range creds {
		cred.Created = time.Now()
		data, err := json.Marshal(cred)
		if err != nil {
			return err
		}
		if err := batch.Set([]byte("Meta:"+cred.ID), data, pebble.Sync); err != nil {
			return err
		}
		if err := batch.Set([]byte("Secret:"+cred.ID), []byte(cred.Notes), pebble.Sync); err != nil {
			return err
		}
	}
	return batch.Commit(pebble.Sync)
}

func (s *VaultPebbleStorage) ListCredentials() ([]Credential, error) {
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte("Meta:"),
		UpperBound: []byte("Meta;"),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var credentials []Credential
	for iter.First(); iter.Valid(); iter.Next() {
		var cred Credential
		if err := json.Unmarshal(iter.Value(), &cred); err != nil {
			return nil, err
		}
		credentials = append(credentials, cred)
	}
	return credentials, nil
}

func (s *VaultPebbleStorage) GetCredential(id string) (*Credential, string, error) {
	metaKey := []byte("Meta:" + id)
	secretKey := []byte("Secret:" + id)

	metaData, closer, err := s.db.Get(metaKey)

	if err != nil {
		return nil, "", err
	}
	defer closer.Close()
	var cred Credential
	if err := json.Unmarshal(metaData, &cred); err != nil {
		return nil, "", err
	}
	encSecret, closer, err := s.db.Get(secretKey)
	log.Printf("[DEBUG] Encrpted Secret len=%d", len(encSecret))
	if err != nil {
		return nil, "", err
	}
	defer closer.Close()
	secretBytes, err := crypto.Decrypt(s.key, encSecret)
	if err != nil {
		return nil, "", err
	}
	return &cred, string(secretBytes), nil
}

func (s *VaultPebbleStorage) Close() error {
	return s.db.Close()
}
