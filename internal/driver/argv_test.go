package driver

import (
	"errors"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/schema"
)

func TestArgvLoadHosts(t *testing.T) {
	cases := []struct {
		Name          string
		ParserFunc    schema.HostParserFunc
		Input         []string
		ExpectedHosts []schema.Host
		ExpectedError error
	}{
		{
			Name: "OK",
			ParserFunc: func(host schema.Host) (schema.Host, error) {
				return host, nil
			},
			Input:         []string{"192.168.1.1", "192.168.1.2"},
			ExpectedHosts: []schema.Host{schema.Host{IP: "192.168.1.1"}, schema.Host{IP: "192.168.1.2"}},
		},
		{
			Name: "OK, verify calling parser func with result OK",
			ParserFunc: func(host schema.Host) (schema.Host, error) {
				return schema.Host{IP: "parsed"}, nil
			},
			Input:         []string{"192.168.1.1:22", "192.168.1.2:22"},
			ExpectedHosts: []schema.Host{schema.Host{IP: "parsed"}, schema.Host{IP: "parsed"}},
		},
		{
			Name: "ERROR, verify calling parser func with result ERROR",
			ParserFunc: func(host schema.Host) (schema.Host, error) {
				return schema.Host{}, errors.New("some error")
			},
			Input:         []string{"192.168.1.1:22", "192.168.1.2:22"},
			ExpectedHosts: nil,
			ExpectedError: errors.New("some error"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			hosts, err := ArgvLoadHosts(tc.ParserFunc, tc.Input)
			if !reflect.DeepEqual(err, tc.ExpectedError) {
				t.Errorf("got:%v, expected:%v", err, tc.ExpectedError)
			}
			if !reflect.DeepEqual(hosts, tc.ExpectedHosts) {
				t.Errorf("got:%v, expected:%v", hosts, tc.ExpectedHosts)
			}
		})
	}
}
