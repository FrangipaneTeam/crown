package db

import (
	"fmt"

	"go.etcd.io/bbolt"
)

var DataBase *DB

type Name string

const (
	DBTrack Name = "Track"
	DBEvent Name = "Event"
)

var DBNames = []Name{
	DBTrack,
	DBEvent,
}

type DB struct {
	*bbolt.DB
}

func NewDB(path string) error {
	db, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return err
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		for _, name := range DBNames {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return fmt.Errorf("could not create %s bucket: %w", name, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	DataBase = &DB{db}
	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

// TrackDB Return the TrackDB.
func TrackDB() *Name {
	x := Name(DBTrack) //nolint:unconvert
	return &x
}

// Set sets the value for the given key.
func (db *Name) Set(key, value []byte) error {
	return DataBase.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(*db))
		return b.Put(key, value)
	})
}

// Get gets the value for the given key.
func (db *Name) Get(key []byte) ([]byte, error) {
	var value []byte
	err := DataBase.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(*db))
		value = b.Get(key)
		return nil
	})
	return value, err
}

// KeyExists checks if the key exists.
func (db *Name) KeyExists(key []byte) bool {
	var value []byte
	err := DataBase.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(*db))
		value = b.Get(key)
		return nil
	})
	return err == nil && value != nil
}

// GetAllKeys return all the keys.
func (db *Name) GetAllKeys() ([]string, error) {
	var keys []string
	err := DataBase.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(*db))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			keys = append(keys, string(k))
		}
		return nil
	})
	return keys, err
}

// Delete deletes the value for the given key.
func (db *Name) Delete(key []byte) error {
	return DataBase.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(*db))
		return b.Delete(key)
	})
}

// Byte Return []byte of the string.
func Byte(s string) []byte {
	return []byte(s)
}

// Client is a client for interacting with the database.
func (db *Name) Client() *bbolt.DB {
	return DataBase.DB
}

// Bucket return the bucket.
func (db *Name) Bucket() string {
	return string(*db)
}
