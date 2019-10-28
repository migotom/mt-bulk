package kvdb

import (
	"github.com/dgraph-io/badger"
)

type KV interface {
	View(fn func(Txn) error) error
	NewTransaction() Txn
	Close() error
}

type Txn interface {
	NewIterator(badger.IteratorOptions) Iterator
	Discard()
	Commit() error
	GetCopy(key string, decoded interface{}) error
	Store(key string, value interface{}) error
}

type Iterator interface {
	Close()
	Rewind()
	Item() Item
	Next()
	Valid() bool
}

type Item interface {
	ValueCopy(dst []byte) ([]byte, error)
	KeyCopy(dst []byte) []byte
}
