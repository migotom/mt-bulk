package driver

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/schema"
)

func TestFileLoadHosts(t *testing.T) {
	cases := []struct {
		Name          string
		ParserFunc    schema.HostParserFunc
		FileContent   string
		ExpectedHosts []schema.Host
		ExpectedError error
	}{
		{
			Name: "OK",
			ParserFunc: func(host schema.Host) (schema.Host, error) {
				return host, nil
			},
			FileContent:   "192.168.1.1\n192.168.1.2",
			ExpectedHosts: []schema.Host{schema.Host{IP: "192.168.1.1"}, schema.Host{IP: "192.168.1.2"}},
		},
		{
			Name: "OK, verify calling parser func with result OK",
			ParserFunc: func(host schema.Host) (schema.Host, error) {
				return schema.Host{IP: "parsed"}, nil
			},
			FileContent:   "192.168.1.1:22\n192.168.1.2:22",
			ExpectedHosts: []schema.Host{schema.Host{IP: "parsed"}, schema.Host{IP: "parsed"}},
		},
		{
			Name: "ERROR, verify calling parser func with result ERROR",
			ParserFunc: func(host schema.Host) (schema.Host, error) {
				return schema.Host{}, errors.New("some error")
			},
			FileContent:   "192.168.1.1:22\n192.168.1.2:22",
			ExpectedHosts: nil,
			ExpectedError: errors.New("some error"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tmpfile, err := ioutil.TempFile("", "test.txt")
			if err != nil {
				t.Errorf("Can't create temporary test file %v", err)
			}
			defer os.Remove(tmpfile.Name())
			if _, err := tmpfile.Write([]byte(tc.FileContent)); err != nil {
				t.Errorf("Can't write to temporary test file %v", err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Errorf("Can't close temporary test file %v", err)
			}

			hosts, err := FileLoadHosts(tc.ParserFunc, tmpfile.Name())
			if !reflect.DeepEqual(err, tc.ExpectedError) {
				t.Errorf("got:%v, expected:%v", err, tc.ExpectedError)
			}
			if !reflect.DeepEqual(hosts, tc.ExpectedHosts) {
				t.Errorf("got:%v, expected:%v", hosts, tc.ExpectedHosts)
			}
		})
	}
}
