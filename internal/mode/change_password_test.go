package mode

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/clients/mocks"
	"github.com/migotom/mt-bulk/internal/entities"
)

func TestChangePassword(t *testing.T) {
	cases := []struct {
		Name          string
		Job           entities.Job
		Expected      []entities.CommandResult
		ExpectedError error
	}{
		{
			Name: "OK",
			Job:  entities.Job{Host: entities.Host{Password: "old"}, Data: map[string]string{"new_password": "secret"}},
			Expected: []entities.CommandResult{
				entities.CommandResult{Body: "/user/set =numbers=admin =password=secret", Responses: []string{"/user/set =numbers=admin =password=secret"}},
			},
		},
		{
			Name:          "Wrong, missing new password",
			Job:           entities.Job{Host: entities.Host{Password: "old"}},
			Expected:      nil,
			ExpectedError: errors.New("missing or empty new password for change password operation"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			client := mocks.Client{}

			results, err := ChangePassword(context.Background(), client, &tc.Job)
			if !reflect.DeepEqual(err, tc.ExpectedError) {
				t.Errorf("got:%v, expected:%v", err, tc.ExpectedError)
			}

			if !reflect.DeepEqual(results, tc.Expected) {
				t.Errorf("not expected commands:%v, expected:%v", results, tc.Expected)
			}
		})
	}
}
