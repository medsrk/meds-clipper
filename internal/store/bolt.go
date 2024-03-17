package store

import (
	"encoding/binary"
	"encoding/json"
	"go.etcd.io/bbolt"
)

type BoltStore struct {
	db *bbolt.DB
}

func NewBoltStore(dbPath string) (*BoltStore, error) {
	// open the database file
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1}) // 0600 only the owner can read and write
	if err != nil {
		return nil, err
	}

	// create the bucket if it doesn't exist
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("clipboard"))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &BoltStore{db: db}, nil
}

func (s *BoltStore) SaveClipboardItem(item string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("clipboard"))
		id, _ := b.NextSequence()
		key := itob(id)
		value, err := json.Marshal(item)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

func (s *BoltStore) GetClipboardHistory() ([]string, error) {
	var history []string
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("clipboard"))
		return b.ForEach(func(k, v []byte) error {
			var item string
			if err := json.Unmarshal(v, &item); err != nil {
				return err
			}
			history = append(history, item)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return history, nil

}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
