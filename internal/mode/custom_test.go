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
		Name     string
		Job      entities.Job
		Expected entities.Result
	}{
		{
			Name: "OK",
			Job:  entities.Job{Host: entities.Host{Password: "old"}, Commands: []entities.Command{entities.Command{Body: `/certificate print detail`}}},
			Expected: entities.Result{Results: []entities.CommandResult{
				entities.CommandResult{Body: "/<mt-bulk>establish connection", Responses: []string{"/<mt-bulk>establish connection", " --> attempt #0, password #0, job #"}},
				entities.CommandResult{Body: `/certificate print detail`, Responses: []string{`/certificate print detail`}},
			},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			sugar := zap.NewExample().Sugar()
			client := mocks.Client{}

			result := Custom(context.Background(), sugar, client, &tc.Job)
			if !reflect.DeepEqual(result, tc.Expected) {
				t.Errorf("got:%v, expected:%v", result, tc.Expected)
			}
		})
	}
}
