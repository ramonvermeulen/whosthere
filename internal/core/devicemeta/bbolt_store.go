package devicemeta

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
	bolterrors "go.etcd.io/bbolt/errors"
)

var ErrInvalidMAC = errors.New("invalid MAC address")

var (
	metaBucket         = []byte("meta")
	devicesBucket      = []byte("devices")
	schemaVersionKey   = []byte("schema_version")
	currentSchemaValue = []byte("1")
)

type boltStore struct {
	db *bolt.DB
}

// Open opens or creates a bbolt-backed device metadata store.
func Open(path string) (Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create metadata dir: %w", err)
	}

	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open metadata store: %w", err)
	}

	store := &boltStore{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *boltStore) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		meta, err := tx.CreateBucketIfNotExists(metaBucket)
		if err != nil {
			return fmt.Errorf("create meta bucket: %w", err)
		}

		if _, err := tx.CreateBucketIfNotExists(devicesBucket); err != nil {
			return fmt.Errorf("create devices bucket: %w", err)
		}

		version := meta.Get(schemaVersionKey)
		if len(version) == 0 {
			return meta.Put(schemaVersionKey, currentSchemaValue)
		}
		if !bytes.Equal(version, currentSchemaValue) {
			return fmt.Errorf("unsupported metadata schema version %q", version)
		}

		return nil
	})
}

func (s *boltStore) Get(mac string) (Record, bool, error) {
	normalized, ok := NormalizeMAC(mac)
	if !ok {
		return Record{}, false, ErrInvalidMAC
	}

	var (
		record Record
		found  bool
	)

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(devicesBucket)
		if bucket == nil {
			return nil
		}

		raw := bucket.Get([]byte(normalized))
		if len(raw) == 0 {
			return nil
		}

		found = true
		if err := json.Unmarshal(raw, &record); err != nil {
			return fmt.Errorf("decode record for %s: %w", normalized, err)
		}

		return nil
	})
	if err != nil {
		return Record{}, false, err
	}

	return record, found, nil
}

func (s *boltStore) SetAlias(mac, alias string) error {
	normalized, ok := NormalizeMAC(mac)
	if !ok {
		return ErrInvalidMAC
	}

	alias = strings.TrimSpace(alias)
	if alias == "" {
		return s.ClearAlias(normalized)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(devicesBucket)
		if bucket == nil {
			return errors.New("devices bucket not initialized")
		}

		record := Record{
			Alias:     alias,
			UpdatedAt: time.Now(),
		}

		raw, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("encode record for %s: %w", normalized, err)
		}

		return bucket.Put([]byte(normalized), raw)
	})
}

func (s *boltStore) ClearAlias(mac string) error {
	normalized, ok := NormalizeMAC(mac)
	if !ok {
		return ErrInvalidMAC
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(devicesBucket)
		if bucket == nil {
			return errors.New("devices bucket not initialized")
		}

		raw := bucket.Get([]byte(normalized))
		if len(raw) == 0 {
			return nil
		}

		var record Record
		if err := json.Unmarshal(raw, &record); err != nil {
			return fmt.Errorf("decode record for %s: %w", normalized, err)
		}

		record.Alias = ""
		record.UpdatedAt = time.Now()
		if record.empty() {
			return bucket.Delete([]byte(normalized))
		}

		updated, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("encode record for %s: %w", normalized, err)
		}

		return bucket.Put([]byte(normalized), updated)
	})
}

func (s *boltStore) ResetAliases() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(devicesBucket); err != nil && !errors.Is(err, bolterrors.ErrBucketNotFound) {
			return fmt.Errorf("delete devices bucket: %w", err)
		}

		if _, err := tx.CreateBucketIfNotExists(devicesBucket); err != nil {
			return fmt.Errorf("recreate devices bucket: %w", err)
		}

		return nil
	})
}

func (s *boltStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.Close()
}
