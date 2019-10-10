package driver

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/entities"
)

func TestFileLoadJobs(t *testing.T) {
	cases := []struct {
		Name          string
		FileContent   string
		ExpectedJobs  []entities.Job
		ExpectedError error
	}{
		{
			Name:        "OK",
			FileContent: "192.168.1.1\n192.168.1.2:44",
			ExpectedJobs: []entities.Job{
				entities.Job{Host: entities.Host{IP: "192.168.1.1"}},
				entities.Job{Host: entities.Host{IP: "192.168.1.2", Port: "44"}},
			},
		},
		{
			Name:          "Wrong, invalid hostname",
			FileContent:   "foo\n192.168.1.2:22",
			ExpectedJobs:  nil,
			ExpectedError: errors.New("can't resolve host: foo"),
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

			hosts, err := FileLoadJobs(context.Background(), entities.Job{}, tmpfile.Name())
			if !reflect.DeepEqual(err, tc.ExpectedError) {
				t.Errorf("got:%v, expected:%v", err, tc.ExpectedError)
			}
			if !reflect.DeepEqual(hosts, tc.ExpectedJobs) {
				t.Errorf("got:%v, expected:%v", hosts, tc.ExpectedJobs)
			}
		})
	}
}
