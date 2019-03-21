package schema

import (
	"errors"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	cases := []struct {
		Name           string
		Hosts          Hosts
		ExpectedResult []Host
	}{
		{
			Name:           "OK",
			Hosts:          Hosts{hosts: []Host{Host{IP: "192.168.1.1"}}},
			ExpectedResult: []Host{Host{IP: "192.168.1.1"}},
		},
		{
			Name:           "Empty",
			Hosts:          Hosts{hosts: nil},
			ExpectedResult: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if !reflect.DeepEqual(tc.Hosts.Get(), tc.ExpectedResult) {
				t.Errorf("got:%v, expected:%v", tc.Hosts.Get(), tc.ExpectedResult)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	cases := []struct {
		Name          string
		Inputs        []string
		ExpectedList  []Host
		ExpectedError error
	}{
		{
			Name:          "OK just IP",
			Inputs:        []string{"192.168.1.1", "192.168.1.2"},
			ExpectedList:  []Host{Host{IP: "192.168.1.1"}, Host{IP: "192.168.1.2"}},
			ExpectedError: nil,
		},
		{
			Name:          "OK IP:port",
			Inputs:        []string{"192.168.1.1:22", "192.168.1.2"},
			ExpectedList:  []Host{Host{IP: "192.168.1.1", Port: "22"}, Host{IP: "192.168.1.2"}},
			ExpectedError: nil,
		},
		{
			Name:          "OK hostname:port",
			Inputs:        []string{"wp.pl:22", "192.168.1.2"},
			ExpectedList:  []Host{Host{IP: "212.77.98.9", Port: "22"}, Host{IP: "192.168.1.2"}},
			ExpectedError: nil,
		},
		{
			Name:          "Invalid IP",
			Inputs:        []string{"192,1268.8.8", "192.168.1.2"},
			ExpectedList:  nil,
			ExpectedError: errors.New("hosts loader add can't resolve host: 192,1268.8.8"),
		},
		{
			Name:          "Invalid Port",
			Inputs:        []string{"192.168.8.8:XX", "192.168.1.2"},
			ExpectedList:  nil,
			ExpectedError: errors.New("hosts loader add port invalid format: XX"),
		},
		{
			Name:          "Empty",
			Inputs:        []string{},
			ExpectedList:  nil,
			ExpectedError: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var hosts Hosts

			err := hosts.Add(func(parser HostParserFunc) ([]Host, error) {
				return ArgvLoadHosts(parser, tc.Inputs)
			})
			if !reflect.DeepEqual(err, tc.ExpectedError) {
				t.Errorf("not expected error:%v, expected:%v", err, tc.ExpectedError)
			}
			if !reflect.DeepEqual(hosts.Get(), tc.ExpectedList) {
				t.Errorf("got:%v, expected:%v", hosts.Get(), tc.ExpectedList)
			}
		})
	}
}

func ArgvLoadHosts(hostParser HostParserFunc, data []string) (hosts []Host, err error) {
	hosts = make([]Host, len(data))
	for i, entry := range data {
		if hosts[i], err = hostParser(Host{IP: entry}); err != nil {
			return nil, err
		}
	}
	return hosts, nil
}
