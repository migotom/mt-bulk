package vulnerabilities

import (
	"bytes"
	"context"
	"encoding/gob"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/kvdb"
	"github.com/migotom/mt-bulk/internal/kvdb/mocks"
)

func TestCheck(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	const response = `
	{
		"data": [
			{
				"Modified": "2017-08-29T01:32:00",
				"cvss": 6.4,
				"id": "CVE-1-2",
				"summary": "Some issue",
				"references": [],
				"vulnerable_configuration": [
					"cpe:2.3:o:mikrotik:routeros:5.15:*:*:*:*:*:*:*"
				]
			}
		]
	}`

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(response))
	}))
	defer func() { testServer.Close() }()

	cases := []struct {
		Name          string
		Version       string
		ExpectedMocks func(kv *mocks.KVMock)
		ExpectedErr   string
	}{
		{
			Name:        "Found vulnerability",
			Version:     "5.12",
			ExpectedErr: "vulnerabilities found: CVE-1 (0.0)",
			ExpectedMocks: func(kv *mocks.KVMock) {
				kv.Txn.On("GetCopy", "DB:LastUpdate", mock.Anything).Return(time.Now(), nil)
				kv.Txn.On("GetCopy", "DB:Version", mock.Anything).Return(1, nil)

				item := &mocks.ItemMock{}
				item.On("KeyCopy", mock.Anything).Return([]byte("Version:61200"))

				kv.Txn.It = mocks.IteratorMock{Items: []kvdb.Item{item}}

				buffer := bytes.NewBuffer(nil)
				_ = gob.NewEncoder(buffer).Encode([]string{"CVE-1"})

				item.On("ValueCopy", mock.Anything).Return(buffer.Bytes(), nil)
				kv.Txn.On("GetCopy", "CVE:CVE-1", mock.Anything).Return(CVE{ID: "CVE-1", Summary: "URGENT"}, nil)
			},
		},
		{
			Name:        "Not found any vulnerability",
			Version:     "5.12",
			ExpectedErr: "",
			ExpectedMocks: func(kv *mocks.KVMock) {
				kv.Txn.On("GetCopy", "DB:LastUpdate", mock.Anything).Return(time.Now(), nil)
				kv.Txn.On("GetCopy", "DB:Version", mock.Anything).Return(1, nil)

				item := &mocks.ItemMock{}
				item.On("KeyCopy", mock.Anything).Return([]byte("Version:11200"))

				kv.Txn.It = mocks.IteratorMock{Items: []kvdb.Item{item}}
			},
		},
		{
			Name:        "Not loaded any vulnerability",
			Version:     "5.12",
			ExpectedErr: "",
			ExpectedMocks: func(kv *mocks.KVMock) {
				kv.Txn.On("GetCopy", "DB:LastUpdate", mock.Anything).Return(time.Now(), nil)
				kv.Txn.On("GetCopy", "DB:Version", mock.Anything).Return(1, nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			kvMock := mocks.KVMock{Txn: mocks.TxnMock{It: mocks.IteratorMock{}}}
			vm := NewManager(sugar, testServer.URL, &kvMock)
			tc.ExpectedMocks(&kvMock)

			ctx, cancel := context.WithCancel(context.Background())
			go func(ctx context.Context) {
				vm.Listen(ctx)
			}(ctx)
			defer cancel()

			err := vm.Check(tc.Version)

			if err != nil {
				assert.EqualError(t, err, tc.ExpectedErr)
			} else {
				assert.Empty(t, tc.ExpectedErr)
			}
			kvMock.Txn.AssertExpectations(t)
		})
	}

}
