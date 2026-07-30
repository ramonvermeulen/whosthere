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
	scopeMetadataBucket  = []byte("scope_metadata")
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
		if _, err := tx.CreateBucketIfNotExists(scopeMetadataBucket); err != nil {
			return fmt.Errorf("create scope metadata bucket: %w", err)
		}

		return nil
	})
}

func (s *boltStore) Get(scopeID, mac string) (Record, bool, error) {
	normalizedMAC, err := normalizedMACKey(mac)
	if err != nil {
		return Record{}, false, ErrInvalidMAC
	}

	var (
		record Record
		found  bool
	)

	err = s.db.View(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		if deviceMetadata == nil {
			return errors.New("device metadata bucket not initialized")
		}

		keys := candidateDeviceKeys(scopeID, normalizedMAC)
		for _, key := range keys {
			rawRecord := deviceMetadata.Get(key)
			if len(rawRecord) == 0 {
				continue
			}

			found = true
			if err := json.Unmarshal(rawRecord, &record); err != nil {
				return fmt.Errorf("decode record for %s: %w", string(key), err)
			}
			record.MAC = normalizedMAC
			if record.ScopeID == "" && len(key) > len(normalizedMAC) {
				record.ScopeID = scopeID
			}
			record.normalize()
			return nil
		}

		return nil
	})
	if err != nil {
		return Record{}, false, err
	}

	return record, found, nil
}

func (s *boltStore) Upsert(record Record) error {
	normalizedMAC, err := normalizedMACKey(record.MAC)
	if err != nil {
		return ErrInvalidMAC
	}

	record.MAC = normalizedMAC
	record.normalize()
	record.UpdatedAt = time.Now()

	return s.db.Update(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		if deviceMetadata == nil {
			return errors.New("device metadata bucket not initialized")
		}

		raw, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("encode record for %s: %w", deviceStorageKey(record.ScopeID, normalizedMAC), err)
		}

		return deviceMetadata.Put([]byte(deviceStorageKey(record.ScopeID, normalizedMAC)), raw)
	})
}

func (s *boltStore) Delete(scopeID, mac string) error {
	normalizedMAC, err := normalizedMACKey(mac)
	if err != nil {
		return ErrInvalidMAC
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		if deviceMetadata == nil {
			return errors.New("device metadata bucket not initialized")
		}

		return deviceMetadata.Delete([]byte(deviceStorageKey(scopeID, normalizedMAC)))
	})
}

func (s *boltStore) ForEach(fn func(Record) error) error {
	if fn == nil {
		return nil
	}

	return s.db.View(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		if deviceMetadata == nil {
			return errors.New("device metadata bucket not initialized")
		}

		return deviceMetadata.ForEach(func(rawKey, rawRecord []byte) error {
			var record Record
			if err := json.Unmarshal(rawRecord, &record); err != nil {
				return fmt.Errorf("decode record for %s: %w", string(rawKey), err)
			}

			scopeID, mac := parseDeviceStorageKey(string(rawKey))
			record.ScopeID = firstNonEmpty(record.ScopeID, scopeID)
			record.MAC = firstNonEmpty(record.MAC, mac)
			record.normalize()

			return fn(record)
		})
	})
}

func (s *boltStore) UpsertScope(scope ScopeRecord) error {
	scope.normalize()
	if scope.ScopeID == "" {
		return nil
	}

	now := time.Now()
	if scope.FirstSeen.IsZero() {
		scope.FirstSeen = now
	}
	if scope.LastSeen.IsZero() {
		scope.LastSeen = now
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		scopeMetadata := tx.Bucket(scopeMetadataBucket)
		if scopeMetadata == nil {
			return errors.New("scope metadata bucket not initialized")
		}

		existingRaw := scopeMetadata.Get([]byte(scope.ScopeID))
		if len(existingRaw) > 0 {
			var existing ScopeRecord
			if err := json.Unmarshal(existingRaw, &existing); err != nil {
				return fmt.Errorf("decode scope record for %s: %w", scope.ScopeID, err)
			}
			if existing.FirstSeen.IsZero() || (!scope.FirstSeen.IsZero() && scope.FirstSeen.Before(existing.FirstSeen)) {
				existing.FirstSeen = scope.FirstSeen
			}
			if scope.LastSeen.After(existing.LastSeen) {
				existing.LastSeen = scope.LastSeen
			}
			if scope.InterfaceName != "" {
				existing.InterfaceName = scope.InterfaceName
			}
			if scope.NetworkCIDR != "" {
				existing.NetworkCIDR = scope.NetworkCIDR
			}
			if scope.GatewayIP != "" {
				existing.GatewayIP = scope.GatewayIP
			}
			if scope.GatewayMAC != "" {
				existing.GatewayMAC = scope.GatewayMAC
			}
			scope = existing
		}

		raw, err := json.Marshal(scope)
		if err != nil {
			return fmt.Errorf("encode scope record for %s: %w", scope.ScopeID, err)
		}

		return scopeMetadata.Put([]byte(scope.ScopeID), raw)
	})
}

func (s *boltStore) GetScope(scopeID string) (ScopeRecord, bool, error) {
	scopeID = strings.TrimSpace(scopeID)
	if scopeID == "" {
		return ScopeRecord{}, false, nil
	}

	var (
		record ScopeRecord
		found  bool
	)

	err := s.db.View(func(tx *bolt.Tx) error {
		scopeMetadata := tx.Bucket(scopeMetadataBucket)
		if scopeMetadata == nil {
			return errors.New("scope metadata bucket not initialized")
		}

		raw := scopeMetadata.Get([]byte(scopeID))
		if len(raw) == 0 {
			return nil
		}

		found = true
		if err := json.Unmarshal(raw, &record); err != nil {
			return fmt.Errorf("decode scope record for %s: %w", scopeID, err)
		}
		record.ScopeID = scopeID
		record.normalize()
		return nil
	})
	if err != nil {
		return ScopeRecord{}, false, err
	}

	return record, found, nil
}

func (s *boltStore) DeleteScopeAndDevices(scopeID string) error {
	scopeID = strings.TrimSpace(scopeID)
	if scopeID == "" {
		return nil
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		scopeMetadata := tx.Bucket(scopeMetadataBucket)
		if deviceMetadata == nil || scopeMetadata == nil {
			return errors.New("metadata buckets not initialized")
		}

		prefix := scopeID + ":"
		var keysToDelete [][]byte
		if err := deviceMetadata.ForEach(func(rawKey, _ []byte) error {
			if strings.HasPrefix(string(rawKey), prefix) {
				keysToDelete = append(keysToDelete, append([]byte(nil), rawKey...))
			}
			return nil
		}); err != nil {
			return fmt.Errorf("list device metadata keys for scope %s: %w", scopeID, err)
		}

		for _, key := range keysToDelete {
			if err := deviceMetadata.Delete(key); err != nil {
				return fmt.Errorf("delete device record %s: %w", string(key), err)
			}
		}

		return scopeMetadata.Delete([]byte(scopeID))
	})
}

func (s *boltStore) DeleteAllScopesAndDevices() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(deviceMetadataBucket); err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
			return fmt.Errorf("delete device metadata bucket: %w", err)
		}
		if err := tx.DeleteBucket(scopeMetadataBucket); err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
			return fmt.Errorf("delete scope metadata bucket: %w", err)
		}
		if _, err := tx.CreateBucketIfNotExists(deviceMetadataBucket); err != nil {
			return fmt.Errorf("recreate device metadata bucket: %w", err)
		}
		if _, err := tx.CreateBucketIfNotExists(scopeMetadataBucket); err != nil {
			return fmt.Errorf("recreate scope metadata bucket: %w", err)
		}
		return nil
	})
}

func (s *boltStore) SetAlias(scopeID, mac, alias string) error {
	scopeID = strings.TrimSpace(scopeID)
	normalizedMAC, err := normalizedMACKey(mac)
	if err != nil {
		return ErrInvalidMAC
	}

	alias = strings.TrimSpace(alias)
	if alias == "" {
		return s.ClearAlias(scopeID, normalizedMAC)
	}

	record, found, err := s.Get(scopeID, normalizedMAC)
	if err != nil {
		return err
	}
	if !found {
		record = Record{ScopeID: scopeID, MAC: normalizedMAC}
	}
	record.Alias = alias
	record.ScopeID = scopeID
	return s.Upsert(record)
}

func (s *boltStore) ClearAlias(scopeID, mac string) error {
	scopeID = strings.TrimSpace(scopeID)
	normalizedMAC, err := normalizedMACKey(mac)
	if err != nil {
		return ErrInvalidMAC
	}

	record, found, err := s.Get(scopeID, normalizedMAC)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	record.Alias = ""
	record.UpdatedAt = time.Now()
	record.normalize()
	if record.empty() {
		return s.Delete(record.ScopeID, normalizedMAC)
	}

	return s.Upsert(record)
}

func (s *boltStore) ResetAliases() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		deviceMetadata := tx.Bucket(deviceMetadataBucket)
		if deviceMetadata == nil {
			return errors.New("device metadata bucket not initialized")
		}

		var keys [][]byte
		if err := deviceMetadata.ForEach(func(rawKey, _ []byte) error {
			keys = append(keys, append([]byte(nil), rawKey...))
			return nil
		}); err != nil {
			return fmt.Errorf("list device metadata keys: %w", err)
		}

		for _, key := range keys {
			rawRecord := deviceMetadata.Get(key)
			if len(rawRecord) == 0 {
				continue
			}

			var record Record
			if err := json.Unmarshal(rawRecord, &record); err != nil {
				return fmt.Errorf("decode record for %s: %w", string(key), err)
			}

			scopeID, mac := parseDeviceStorageKey(string(key))
			record.ScopeID = firstNonEmpty(record.ScopeID, scopeID)
			record.MAC = firstNonEmpty(record.MAC, mac)
			record.Alias = ""
			record.UpdatedAt = time.Now()
			record.normalize()

			if record.empty() {
				if err := deviceMetadata.Delete(key); err != nil {
					return fmt.Errorf("delete alias for %s: %w", string(key), err)
				}
				continue
			}

			updatedRecord, err := json.Marshal(record)
			if err != nil {
				return fmt.Errorf("encode record for %s: %w", string(key), err)
			}
			if err := deviceMetadata.Put(key, updatedRecord); err != nil {
				return fmt.Errorf("reset alias for %s: %w", string(key), err)
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

func normalizedMACKey(mac string) (string, error) {
	normalizedMAC, isValidMAC := NormalizeMAC(mac)
	if !isValidMAC {
		return "", ErrInvalidMAC
	}
	return normalizedMAC, nil
}

func deviceStorageKey(scopeID, normalizedMAC string) string {
	scopeID = strings.TrimSpace(scopeID)
	if scopeID == "" {
		return normalizedMAC
	}
	return scopeID + "|" + normalizedMAC
}

func candidateDeviceKeys(scopeID, normalizedMAC string) [][]byte {
	keys := [][]byte{}
	if trimmedScopeID := strings.TrimSpace(scopeID); trimmedScopeID != "" {
		keys = append(keys, []byte(deviceStorageKey(trimmedScopeID, normalizedMAC)))
	}
	keys = append(keys, []byte(normalizedMAC))
	return keys
}

func parseDeviceStorageKey(key string) (scopeID, mac string) {
	if idx := strings.IndexByte(key, '|'); idx >= 0 {
		return key[:idx], key[idx+1:]
	}
	return "", key
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
