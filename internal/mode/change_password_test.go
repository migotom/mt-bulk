package mode

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/clients/mocks"
	"github.com/migotom/mt-bulk/internal/entities"
)

func TestChangePassword(t *testing.T) {
	cases := []struct {
		Name     string
		Job      entities.Job
		Expected entities.Result
	}{
		{
			Name: "OK",
			Job:  entities.Job{Host: entities.Host{Password: "old"}, Data: map[string]string{"new_password": "secret"}},
			Expected: entities.Result{
				Results: []entities.CommandResult{
					entities.CommandResult{Body: "/<mt-bulk>establish connection", Responses: []string{"/<mt-bulk>establish connection", " --> attempt #0, password #0, job #"}},
					entities.CommandResult{Body: "/user/set =numbers=admin =password=secret", Responses: []string{"/user/set =numbers=admin =password=secret"}},
				},
			},
		},
		{
			Name: "Wrong, missing new password",
			Job:  entities.Job{Host: entities.Host{Password: "old"}},
			Expected: entities.Result{
				Errors: []error{errors.New("missing or empty new password for change password operation")},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			sugar := zap.NewExample().Sugar()
			client := mocks.Client{}

			result := ChangePassword(context.Background(), sugar, client, &tc.Job)
			if !reflect.DeepEqual(result, tc.Expected) {
				t.Errorf("got:%v, expected:%v", result, tc.Expected)
			}
		})
	}
}
