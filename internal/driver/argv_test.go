package driver

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/entities"
)

func TestArgvLoadJobs(t *testing.T) {
	cases := []struct {
		Name          string
		Args          []string
		ExpectedJobs  []entities.Job
		ExpectedError error
	}{
		{
			Name: "OK",
			Args: []string{"192.168.1.1", "192.168.1.2:44"},
			ExpectedJobs: []entities.Job{
				entities.Job{Host: entities.Host{IP: "192.168.1.1"}},
				entities.Job{Host: entities.Host{IP: "192.168.1.2", Port: "44"}},
			},
		},
		{
			Name:          "Wrong hostname",
			Args:          []string{"fuu", "bar"},
			ExpectedJobs:  nil,
			ExpectedError: errors.New("can't resolve host: fuu"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			jobs, err := ArgvLoadJobs(context.Background(), entities.Job{}, tc.Args)
			if !reflect.DeepEqual(err, tc.ExpectedError) {
				t.Errorf("got:%v, expected:%v", err, tc.ExpectedError)
			}
			if !reflect.DeepEqual(jobs, tc.ExpectedJobs) {
				t.Errorf("got:%+v, expected:%+v", jobs, tc.ExpectedJobs)
			}
		})
	}
}
