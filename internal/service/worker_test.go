package service

import (
	"testing"

	"github.com/migotom/mt-bulk/internal/entities"
)

func TestWorkerProcessingHost(t *testing.T) {
	cases := []struct {
		Name     string
		Worker   Worker
		Host     entities.Host
		Expected bool
	}{
		{
			Name: "OK, already processing",
			Worker: Worker{processingHosts: []entities.Host{
				entities.Host{IP: "10.0.0.1"},
			}},
			Host:     entities.Host{IP: "10.0.0.1"},
			Expected: true,
		},
		{
			Name: "OK, not processing",
			Worker: Worker{processingHosts: []entities.Host{
				entities.Host{IP: "10.0.0.2"},
			}},
			Host:     entities.Host{IP: "10.0.0.1"},
			Expected: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			processing := tc.Worker.ProcessingHost(tc.Host)
			if processing != tc.Expected {
				t.Errorf("not expected %+v, expecting %+v", processing, tc.Expected)
			}
		})
	}
}
