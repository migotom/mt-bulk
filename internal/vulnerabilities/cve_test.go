package vulnerabilities

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/kvdb/mocks"
	"github.com/stretchr/testify/mock"
)

func TestCVEsDownload(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	type response struct {
		Status int
		Body   string
	}
	cases := []struct {
		Name          string
		Response      response
		ExpectedErr   error
		ExpectedMocks func(kv *mocks.KVMock)
	}{
		{
			Name: "Valid flow",
			Response: response{
				http.StatusOK,
				`
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
					},
					{
						"Modified": "2018-08-29T01:32:00",
						"cvss": 1.4,
						"id": "CVE-2-2",
						"summary": "Some issue no 2",
						"references": [],
						"vulnerable_configuration": [
							"cpe:2.3:o:mikrotik:routeros:2.15:*:*:*:*:*:*:*",
							"cpe:2.3:o:mikrotik:routeros:5.15:*:*:*:*:*:*:*"
						]
					}
				]
			}`,
			},
			ExpectedErr: nil,
			ExpectedMocks: func(kv *mocks.KVMock) {
				kv.Txn.On("Store", "CVE:CVE-1-2", CVE{ID: "CVE-1-2", CVSS: 6.4, Modified: "2017-08-29T01:32:00", Summary: "Some issue", References: []string{}}).Return(nil)
				kv.Txn.On("Store", "CVE:CVE-2-2", CVE{ID: "CVE-2-2", CVSS: 1.4, Modified: "2018-08-29T01:32:00", Summary: "Some issue no 2", References: []string{}}).Return(nil)
				kv.Txn.On("Store", "Version:51500", []string{"CVE-1-2", "CVE-2-2"}).Return(nil)
				kv.Txn.On("Store", "Version:21500", []string{"CVE-2-2"}).Return(nil)
				kv.Txn.On("Store", "DB:LastUpdate", mock.Anything).Return(nil)
				kv.Txn.On("Store", "DB:Version", 1).Return(nil)
				kv.Txn.On("Commit").Return(nil)
				kv.Txn.On("Discard").Return()
			},
		},
		{
			Name: "Error diring storing on KV store",
			Response: response{
				http.StatusOK,
				`
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
			}`,
			},
			ExpectedErr: errors.New("wrong"),
			ExpectedMocks: func(kv *mocks.KVMock) {
				kv.Txn.On("Store", "CVE:CVE-1-2", CVE{ID: "CVE-1-2", CVSS: 6.4, Modified: "2017-08-29T01:32:00", Summary: "Some issue", References: []string{}}).Return(nil)
				kv.Txn.On("Store", "Version:51500", []string{"CVE-1-2"}).Return(errors.New("wrong"))
				kv.Txn.On("Discard").Return()
			},
		},
		{
			Name: "Corrupted response",
			Response: response{
				http.StatusOK,
				`
			{
				Random data 9723972349
			}`,
			},
			ExpectedErr:   errors.New("invalid response body invalid character 'R'"),
			ExpectedMocks: func(kv *mocks.KVMock) {},
		},
		{
			Name: "Corrupted response",
			Response: response{
				http.StatusBadRequest,
				`Bad Request`,
			},
			ExpectedErr:   errors.New("invalid response status code 400"),
			ExpectedMocks: func(kv *mocks.KVMock) {},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tc.Response.Status)
				res.Write([]byte(tc.Response.Body))
			}))
			defer func() { testServer.Close() }()

			kvMock := mocks.KVMock{Txn: mocks.TxnMock{}}
			vm := NewManager(sugar, testServer.URL, &kvMock)

			tc.ExpectedMocks(&kvMock)

			err := vm.CVEsDownload(context.Background())
			if !reflect.DeepEqual(err, tc.ExpectedErr) {
				t.Errorf("not expected error %+v, expecting %+v", err, tc.ExpectedErr)
			}

			kvMock.Txn.AssertExpectations(t)
		})
	}
}
