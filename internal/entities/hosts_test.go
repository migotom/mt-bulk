package entities

import (
	"errors"
	"reflect"
	"testing"
)

func TestGetPasswords(t *testing.T) {
	cases := []struct {
		Name     string
		Host     Host
		Expected []string
	}{
		{
			Name:     "OK",
			Host:     Host{Password: "secret, secret1,  secret2  "},
			Expected: []string{"secret", "secret1", "secret2"},
		},
		{
			Name:     "Only one",
			Host:     Host{Password: "secret  "},
			Expected: []string{"secret"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if !reflect.DeepEqual(tc.Host.GetPasswords(), tc.Expected) {
				t.Errorf("got:%v, expected:%v", tc.Host.GetPasswords(), tc.Expected)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	cases := []struct {
		Name            string
		Host            Host
		DefaultPort     string
		DefaultUser     string
		DefaultPassword string
		Expected        Host
	}{
		{
			Name:            "OK, not replaced",
			Host:            Host{IP: "10.0.0.1", Port: "22", User: "admin", Password: "secret"},
			DefaultPort:     "8080",
			DefaultUser:     "root",
			DefaultPassword: "secret2",
			Expected:        Host{IP: "10.0.0.1", Port: "22", User: "admin", Password: "secret"},
		},
		{
			Name:            "OK, defaults set",
			Host:            Host{IP: "10.0.0.1"},
			DefaultPort:     "8080",
			DefaultUser:     "root",
			DefaultPassword: "secret2",
			Expected:        Host{IP: "10.0.0.1", Port: "8080", User: "root", Password: "secret2"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Host.SetDefaults(tc.DefaultPort, tc.DefaultUser, tc.DefaultPassword)
			if !reflect.DeepEqual(tc.Host, tc.Expected) {
				t.Errorf("got:%v, expected:%v", tc.Host.GetPasswords(), tc.Expected)
			}
		})
	}
}
func TestParse(t *testing.T) {
	cases := []struct {
		Name          string
		Host          Host
		Expected      Host
		ExpectedError error
	}{
		{
			Name:          "OK just IP",
			Host:          Host{IP: "192.168.1.1"},
			Expected:      Host{IP: "192.168.1.1"},
			ExpectedError: nil,
		},
		{
			Name:          "OK IP:port",
			Host:          Host{IP: "192.168.1.1:22"},
			Expected:      Host{IP: "192.168.1.1", Port: "22"},
			ExpectedError: nil,
		},
		{
			Name:          "OK hostname:port",
			Host:          Host{IP: "wp.pl:22"},
			Expected:      Host{IP: "212.77.98.9", Port: "22"},
			ExpectedError: nil,
		},
		{
			Name:          "Invalid IP",
			Host:          Host{IP: "192,1268.8.8"},
			Expected:      Host{IP: "192,1268.8.8"},
			ExpectedError: errors.New("can't resolve host: 192,1268.8.8"),
		},
		{
			Name:          "Invalid Port",
			Host:          Host{IP: "192.168.8.8:XX"},
			Expected:      Host{IP: "192.168.8.8:XX"},
			ExpectedError: errors.New("port invalid format: XX"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Host.Parse()
			if !reflect.DeepEqual(err, tc.ExpectedError) {
				t.Errorf("not expected error:%v, expected:%v", err, tc.ExpectedError)
			}
			if !reflect.DeepEqual(tc.Host, tc.Expected) {
				t.Errorf("got:%v, expected:%v", tc.Host, tc.Expected)
			}
		})
	}
}
