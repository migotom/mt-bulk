package vulnerabilities

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

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
		Name           string
		DBResponse     response
		DBInfoResponse response
		ExpectedErr    error
		ExpectedMocks  func(kv *mocks.KVMock)
	}{
		{
			Name: "Valid flow",
			DBResponse: response{
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
				kv.Txn.On("GetCopy", "DB:CVE:DBInfo", mock.Anything).Return(cveDBInfo{CAPEC: cveDBInfoEntry{cveTime{time.Date(2016, time.October, 28, 17, 22, 15, 0, time.UTC)}}}, nil)
				kv.Txn.On("Store", "DB:CVE:DBInfo", cveDBInfo{CAPEC: cveDBInfoEntry{cveTime{time.Date(2016, time.October, 28, 17, 22, 15, 0, time.UTC)}}}).Return(nil)
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
			DBResponse: response{
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
				kv.Txn.On("GetCopy", "DB:CVE:DBInfo", mock.Anything).Return(cveDBInfo{}, nil)

				kv.Txn.On("Store", "CVE:CVE-1-2", CVE{ID: "CVE-1-2", CVSS: 6.4, Modified: "2017-08-29T01:32:00", Summary: "Some issue", References: []string{}}).Return(nil)
				kv.Txn.On("Store", "Version:51500", []string{"CVE-1-2"}).Return(errors.New("wrong"))
				kv.Txn.On("Discard").Return()
			},
		},
		{
			Name: "Corrupted response",
			DBResponse: response{
				http.StatusOK,
				`
			{
				Random data 9723972349
			}`,
			},
			ExpectedErr: errors.New("invalid response body invalid character 'R'"),
			ExpectedMocks: func(kv *mocks.KVMock) {
				kv.Txn.On("GetCopy", "DB:CVE:DBInfo", mock.Anything).Return(cveDBInfo{}, nil)
				kv.Txn.On("Discard").Return()
			},
		},
		{
			Name: "Corrupted response",
			DBResponse: response{
				http.StatusBadRequest,
				`Bad Request`,
			},
			ExpectedErr: errors.New("invalid response status code 400"),
			ExpectedMocks: func(kv *mocks.KVMock) {
				kv.Txn.On("Discard").Return()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			testServerDBInfo := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tc.DBResponse.Status)
				res.Write([]byte(`
				{
					"capec": {
						"last_update": "2016-10-28T17:22:15",
						"size": 463
					}
				}
				`))
			}))
			defer func() { testServerDBInfo.Close() }()

			testServerDB := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tc.DBResponse.Status)
				res.Write([]byte(tc.DBResponse.Body))
			}))
			defer func() { testServerDB.Close() }()

			kvMock := mocks.KVMock{Txn: mocks.TxnMock{}}
			vm := NewManager(sugar, []CVEURLs{CVEURLs{DBInfo: testServerDBInfo.URL, DB: testServerDB.URL}}, &kvMock)

			tc.ExpectedMocks(&kvMock)
			err := vm.CVEsDownload(context.Background())
			if !reflect.DeepEqual(err, tc.ExpectedErr) {
				t.Errorf("not expected error %+v, expecting %+v", err, tc.ExpectedErr)
			}

			kvMock.Txn.AssertExpectations(t)
		})
	}
}
