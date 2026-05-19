package devicemeta

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var ErrInvalidMAC = errors.New("invalid MAC address")

var (
	deviceMetadataBucket = []byte("device_metadata")
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
		if _, err := tx.CreateBucketIfNotExists(deviceMetadataBucket); err != nil {
			return fmt.Errorf("create device metadata bucket: %w", err)
		}

		return nil
	})
}

func (s *boltStore) Get(mac string) (Record, bool, error) {
	normalizedMAC, isValidMAC := NormalizeMAC(mac)
	if !isValidMAC {
		return Record{}, false, ErrInvalidMAC
	}
	macKey := []byte(normalizedMAC)

	var (
		record Record
		found  bool
	)

	err := s.db.View(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		if deviceMetadata == nil {
			return errors.New("device metadata bucket not initialized")
		}

		rawRecord := deviceMetadata.Get(macKey)
		if len(rawRecord) == 0 {
			return nil
		}

		found = true
		if err := json.Unmarshal(rawRecord, &record); err != nil {
			return fmt.Errorf("decode record for %s: %w", normalizedMAC, err)
		}

		return nil
	})
	if err != nil {
		return Record{}, false, err
	}

	return record, found, nil
}

func (s *boltStore) SetAlias(mac, alias string) error {
	normalizedMAC, isValidMAC := NormalizeMAC(mac)
	if !isValidMAC {
		return ErrInvalidMAC
	}
	macKey := []byte(normalizedMAC)

	alias = strings.TrimSpace(alias)
	if alias == "" {
		return s.ClearAlias(normalizedMAC)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		if deviceMetadata == nil {
			return errors.New("device metadata bucket not initialized")
		}

		record := Record{
			Alias:     alias,
			UpdatedAt: time.Now(),
		}

		raw, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("encode record for %s: %w", normalizedMAC, err)
		}

		return deviceMetadata.Put(macKey, raw)
	})
}

func (s *boltStore) ClearAlias(mac string) error {
	normalizedMAC, isValidMAC := NormalizeMAC(mac)
	if !isValidMAC {
		return ErrInvalidMAC
	}
	macKey := []byte(normalizedMAC)

	return s.db.Update(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		if deviceMetadata == nil {
			return errors.New("device metadata bucket not initialized")
		}

		rawRecord := deviceMetadata.Get(macKey)
		if len(rawRecord) == 0 {
			return nil
		}

		var record Record
		if err := json.Unmarshal(rawRecord, &record); err != nil {
			return fmt.Errorf("decode record for %s: %w", normalizedMAC, err)
		}

		record.Alias = ""
		record.UpdatedAt = time.Now()
		if record.empty() {
			return deviceMetadata.Delete(macKey)
		}

		updatedRecord, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("encode record for %s: %w", normalizedMAC, err)
		}

		return deviceMetadata.Put(macKey, updatedRecord)
	})
}

func (s *boltStore) ResetAliases() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		if deviceMetadata == nil {
			return errors.New("device metadata bucket not initialized")
		}

		var macKeys [][]byte
		if err := deviceMetadata.ForEach(func(macKey, _ []byte) error {
			macKeyCopy := append([]byte(nil), macKey...)
			macKeys = append(macKeys, macKeyCopy)
			return nil
		}); err != nil {
			return fmt.Errorf("list device metadata keys: %w", err)
		}

		for _, macKey := range macKeys {
			if err := deviceMetadata.Delete(macKey); err != nil {
				return fmt.Errorf("delete alias for %s: %w", string(macKey), err)
			}
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
