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
			Name: "OK, use again already processing this host",
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
				entities.Host{IP: "10.0.0.1"},
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			wp := NewWorkerPool(2)
			for _, w := range tc.Workers {
				wx := w
				wp.Add(&wx)
			}

			freshWorker := wp.Get(tc.Host)
			if !reflect.DeepEqual(freshWorker, &tc.Expected) {
				t.Errorf("not expected %+v, expecting %+v", freshWorker, tc.Expected)
			}
		})
	}
}

func TestWorkerPoolGetMultiple(t *testing.T) {
	cases := []struct {
		Name     string
		Workers  []Worker
		Host     entities.Host
		Expected []Worker
	}{
		{
			Name: "OK, pick twice same host already processing this host",
			Workers: []Worker{
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.1"},
				}},
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.2"},
				}},
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.3"},
				}},
			},
			Host: entities.Host{IP: "10.0.0.2"},
			Expected: []Worker{
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.2"}}},
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.2"}}},
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.2"}}},
			},
		},
		{
			Name: "OK, omit already used",
			Workers: []Worker{
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.1"},
				}},
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.2"},
				}},
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.3"},
				}},
			},
			Host: entities.Host{IP: "10.0.0.10"},
			Expected: []Worker{
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.2"}}},
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.3"}}},
				Worker{processingHosts: []entities.Host{
					entities.Host{IP: "10.0.0.1"}}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			wp := NewWorkerPool(3)
			for _, w := range tc.Workers {
				wx := w
				wp.Add(&wx)
			}

			freshWorker := wp.Get(tc.Host)
			if !reflect.DeepEqual(freshWorker, &tc.Expected[0]) {
				t.Errorf("not expected %+v, expecting %+v", freshWorker, tc.Expected)
			}

			freshWorker = wp.Get(tc.Host)
			if !reflect.DeepEqual(freshWorker, &tc.Expected[1]) {
				t.Errorf("not expected %+v, expecting %+v", freshWorker, tc.Expected)
			}

			freshWorker = wp.Get(tc.Host)
			if !reflect.DeepEqual(freshWorker, &tc.Expected[2]) {
				t.Errorf("not expected %+v, expecting %+v", freshWorker, tc.Expected)
			}
		})
	}
}
