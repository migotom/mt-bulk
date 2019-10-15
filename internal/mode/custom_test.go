package mode

import (
	"context"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/clients/mocks"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

func TestCustom(t *testing.T) {
	cases := []struct {
		Name          string
		Job           entities.Job
		Expected      []entities.CommandResult
		ExpectedError error
	}{
		{
			Name: "OK",
			Job:  entities.Job{Host: entities.Host{Password: "old"}, Commands: []entities.Command{entities.Command{Body: `/certificate print detail`}}},
			Expected: []entities.CommandResult{
				entities.CommandResult{Body: "/<mt-bulk>establish connection", Responses: []string{"/<mt-bulk>establish connection", " --> attempt #0, password #0, job #"}},
				entities.CommandResult{Body: `/certificate print detail`, Responses: []string{`/certificate print detail`}},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			sugar := zap.NewExample().Sugar()
			client := mocks.Client{}

			results, err := Custom(context.Background(), sugar, client, &tc.Job)
			if !reflect.DeepEqual(err, tc.ExpectedError) {
				t.Errorf("got:%v, expected:%v", err, tc.ExpectedError)
			}

			if !reflect.DeepEqual(results, tc.Expected) {
				t.Errorf("not expected commands %v", results)
			}
		})
	}
}
