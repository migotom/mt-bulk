package mode

import (
	"context"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/clients/mocks"
	"github.com/migotom/mt-bulk/internal/entities"
)

func TestCustom(t *testing.T) {
	cases := []struct {
		Name          string
		Job           entities.Job
		Expected      []string
		ExpectedError error
	}{
		{
			Name: "OK",
			Job:  entities.Job{Host: entities.Host{Password: "old"}, Commands: []entities.Command{entities.Command{Body: `/certificate print detail`}}},
			Expected: []string{
				`/certificate print detail`,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			client := mocks.Client{}

			results, err := Custom(context.Background(), client, &tc.Job)
			if !reflect.DeepEqual(err, tc.ExpectedError) {
				t.Errorf("got:%v, expected:%v", err, tc.ExpectedError)
			}

			if !reflect.DeepEqual(results, tc.Expected) {
				t.Errorf("not expected commands %v", results)
			}
		})
	}
}
