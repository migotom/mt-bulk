package service

import (
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/entities"
)

func TestWorkerPoolGet(t *testing.T) {
	cases := []struct {
		Name     string
		Workers  []Worker
		Host     entities.Host
		Expected Worker
	}{
		{
			Name: "OK, only one and available",
			Workers: []Worker{
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.1"},
				}},
			},
			Host: entities.Host{IP: "10.0.0.2"},
			Expected: Worker{processingHosts: []entities.Host{
				entities.Host{IP: "10.0.0.1"},
			}},
		},
		{
			Name: "OK, ommit already used",
			Workers: []Worker{
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.1"},
				}},
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.2"},
				}},
			},
			Host: entities.Host{IP: "10.0.0.1"},
			Expected: Worker{processingHosts: []entities.Host{
				entities.Host{IP: "10.0.0.2"},
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			wp := NewWorkerPool(1)
			for _, w := range tc.Workers {
				wp.Add(&w)
			}

			freshWorker := wp.Get(tc.Host)
			if !reflect.DeepEqual(freshWorker, &tc.Expected) {
				t.Errorf("not expected %+v, expecting %+v", freshWorker, tc.Expected)
			}
		})
	}
}
