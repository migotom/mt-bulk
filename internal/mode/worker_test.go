package mode

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service/mocks"
)

func TestWorker(t *testing.T) {
	cases := []struct {
		Name     string
		Hosts    []schema.Host
		Handler  schema.ModeHandlerFunc
		Finisher func(chan schema.Host, chan schema.Error, context.CancelFunc)
	}{
		{
			Name:  "OK",
			Hosts: []schema.Host{schema.Host{IP: "192.168.1.1"}},
			Handler: func(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
				return nil
			},
			Finisher: func(h chan schema.Host, e chan schema.Error, cancel context.CancelFunc) {
				close(h)
			},
		},
		{
			Name:  "Cancel",
			Hosts: []schema.Host{schema.Host{IP: "192.168.1.1"}},
			Handler: func(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
				return nil
			},
			Finisher: func(h chan schema.Host, e chan schema.Error, cancel context.CancelFunc) {
				cancel()
			},
		},
		{
			Name:  "ErrorHandle",
			Hosts: []schema.Host{schema.Host{IP: "192.168.1.1"}},
			Handler: func(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
				return errors.New("some error")
			},
			Finisher: func(h chan schema.Host, e chan schema.Error, cancel context.CancelFunc) {
				if err := <-e; !reflect.DeepEqual(err, schema.Error{Host: schema.Host{IP: "192.168.1.1"}, Message: "some error"}) {
					t.Errorf("not expected error %v", err)
				}
				close(h)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			Service := mocks.Service{}
			appConfig := schema.GeneralConfig{ModeHandler: tc.Handler, IgnoreErrors: true}
			appConfig.Service = make(map[string]*schema.Service)
			appConfig.Service["ssh"] = &schema.Service{
				Interface: &Service,
			}

			ctx, cancel := context.WithCancel(context.Background())

			wg := new(sync.WaitGroup)
			hostsChan := make(chan schema.Host)
			resultsChan := make(chan schema.Error)

			wg.Add(1)
			go Worker(ctx, &appConfig, hostsChan, resultsChan, wg)

			for _, host := range tc.Hosts {
				hostsChan <- host
			}
			tc.Finisher(hostsChan, resultsChan, cancel)
			wg.Wait()

		})
	}
}
