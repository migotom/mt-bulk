package kvdb

import (
	"bytes"
	"encoding/gob"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
	"go.uber.org/zap"
)

// OpenKV opens badger key/value database.
func OpenKV(sugar *zap.SugaredLogger, dbDir string) (KV, error) {
	opts := badger.DefaultOptions(dbDir).WithTruncate(true)
	opts.ValueLogLoadingMode = options.FileIO
	opts.Logger = &dbLog{Sugar: sugar}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &KVDB{
		DB: db,
	}, nil
}

// KVDB is Key/Value internal store.
type KVDB struct {
	*badger.DB
}

// Close cloes badger key/value database.
func (kvdb *KVDB) Close() error {
	defer kvdb.DB.Close()

	return kvdb.DB.Flatten(4)
}

// View creates new view transaction.
func (kvdb *KVDB) View(fn func(Txn) error) error {
	txn := kvdb.NewTransaction()
	defer txn.Discard()

	return fn(txn)
}

// NewTransaction returns new badger transaction.
func (kvdb *KVDB) NewTransaction() Txn {
	return &KVDBTxn{txn: kvdb.DB.NewTransaction(true)}
}

// Txn is an wrapper to badger transaction.
type KVDBTxn struct {
	txn *badger.Txn
}

func (t *KVDBTxn) Commit() error {
	return t.txn.Commit()
}

func (t *KVDBTxn) Discard() {
	t.txn.Discard()
}

func (t *KVDBTxn) NewIterator(options badger.IteratorOptions) Iterator {
	return &KVDBIterator{it: t.txn.NewIterator(options)}
}

// GetCopy gets copy of data from badger database on given key location.
func (t *KVDBTxn) GetCopy(key string, decoded interface{}) error {
	item, err := t.txn.Get([]byte(key))
	if err != nil {
		return err
	}
	value, err := item.ValueCopy(nil)
	if err != nil {
		return err
	}
	err = gob.NewDecoder(bytes.NewReader(value)).Decode(decoded)
	if err != nil {
		return err
	}
	return nil
}

// Store stores value on key location in badger database.
func (t *KVDBTxn) Store(key string, value interface{}) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(value)
	if err != nil {
		return err
	}

	err = t.txn.SetEntry(badger.NewEntry([]byte(key), buf.Bytes()))
	if err != nil {
		return err
	}
	return nil
}

type KVDBIterator struct {
	it *badger.Iterator
}

func (i *KVDBIterator) Close() {
	i.it.Close()
}

func (i *KVDBIterator) Item() Item {
	return i.it.Item()
}

func (i *KVDBIterator) Next() {
	i.it.Next()
}

func (i *KVDBIterator) Rewind() {
	i.it.Rewind()
}

func (i *KVDBIterator) Valid() bool {
	return i.it.Valid()
}

type KVDBItem struct {
	item *badger.Item
}

func (i *KVDBItem) ValueCopy(dst []byte) ([]byte, error) {
	return i.item.ValueCopy(dst)
}

func (i *KVDBItem) KeyCopy(dst []byte) []byte {
	return i.item.KeyCopy(dst)
}
