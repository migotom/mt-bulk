package mocks

import (
	"reflect"

	"github.com/migotom/mt-bulk/internal/kvdb"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/mock"
)

// KVMock mocks KV.
type KVMock struct {
	Txn TxnMock
	mock.Mock
}

// View implements KV's View.
func (kv *KVMock) View(fn func(kvdb.Txn) error) error {
	return fn(kv.NewTransaction())
}

// NewTransaction implements KV's NewTransaction.
func (kv *KVMock) NewTransaction() kvdb.Txn {
	return &kv.Txn
}

// Close implements KV's Close.
func (kv *KVMock) Close() error {
	return nil
}

// TxnMock implements Txn.
type TxnMock struct {
	It IteratorMock
	mock.Mock
}

// NewIterator implements Txn's NewIterator.
func (txn *TxnMock) NewIterator(badger.IteratorOptions) kvdb.Iterator {
	return &txn.It
}

// Discard implements Txn's Discard.
func (txn *TxnMock) Discard() {
	_ = txn.Called()
}

// Commit implements Txn's Commit.
func (txn *TxnMock) Commit() error {
	arg := txn.Called()
	return arg.Error(0)
}

// GetCopy implements Txn's GetCopy.
func (txn *TxnMock) GetCopy(key string, decoded interface{}) error {
	arg := txn.Called(key, decoded)
	val := reflect.ValueOf(decoded)
	val.Elem().Set(reflect.ValueOf(arg.Get(0)))
	return arg.Error(1)
}

// Store implements Txn's Store.
func (txn *TxnMock) Store(key string, value interface{}) error {
	arg := txn.Called(key, value)
	return arg.Error(0)
}

// IteratorMock implements Iterator.
type IteratorMock struct {
	Items []kvdb.Item
	cur   int

	mock.Mock
}

// Close implements Iterator's Close.
func (it *IteratorMock) Close() {}

// Rewind implements Iterator's Rewind.
func (it *IteratorMock) Rewind() {
	it.cur = 0
}

// Item implements Iterator's Item.
func (it *IteratorMock) Item() kvdb.Item {
	return it.Items[it.cur]
}

// Next implements Iterator's Next.
func (it *IteratorMock) Next() {
	it.cur++
}

// Valid implements Iterator's Valid.
func (it *IteratorMock) Valid() bool {
	return it.cur < len(it.Items)
}

// ItemMock implements Item.
type ItemMock struct {
	mock.Mock
}

// ValueCopy implements Item's ValueCopy.
func (i *ItemMock) ValueCopy(dst []byte) ([]byte, error) {
	arg := i.Called(dst)
	return arg.Get(0).([]byte), arg.Error(1)
}

// KeyCopy implements Item's KeyCopy.
func (i *ItemMock) KeyCopy(dst []byte) []byte {
	arg := i.Called(dst)
	return arg.Get(0).([]byte)
}
