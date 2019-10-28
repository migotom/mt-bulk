package mode

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/clients/mocks"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

func TestInitSecureAPI(t *testing.T) {
	cases := []struct {
		Name          string
		Job           entities.Job
		Expected      entities.Result
		ExpectedError error
	}{
		{
			Name: "OK",
			Job:  entities.Job{Host: entities.Host{Password: "old"}, Data: map[string]string{"keys_directory": "certs/"}},
			Expected: entities.Result{Results: []entities.CommandResult{
				entities.CommandResult{Body: "/<mt-bulk>establish connection", Responses: []string{"/<mt-bulk>establish connection", " --> attempt #0, password #0, job #"}},
				entities.CommandResult{Body: `/<mt-bulk>copy sftp://` + filepath.FromSlash("certs/device.crt") + ` sftp://mtbulkdevice.crt`},
				entities.CommandResult{Body: `/<mt-bulk>copy sftp://` + filepath.FromSlash("certs/device.key") + ` sftp://mtbulkdevice.key`},
				entities.CommandResult{Body: `/ip service set api-ssl certificate=none`, Responses: []string{`/ip service set api-ssl certificate=none`}},
				entities.CommandResult{Body: `/certificate print detail`, Responses: []string{`/certificate print detail`}},
				entities.CommandResult{Body: `/certificate remove %{c1}`, Responses: []string{`/certificate remove %{c1}`}},
				entities.CommandResult{Body: `/certificate import file-name=mtbulkdevice.crt passphrase=""`, Responses: []string{`/certificate import file-name=mtbulkdevice.crt passphrase=""`}},
				entities.CommandResult{Body: `/certificate import file-name=mtbulkdevice.key passphrase=""`, Responses: []string{`/certificate import file-name=mtbulkdevice.key passphrase=""`}},
				entities.CommandResult{Body: `/ip service set api-ssl disabled=no certificate=mtbulkdevice.crt`, Responses: []string{`/ip service set api-ssl disabled=no certificate=mtbulkdevice.crt`}},
			},
			},
		},
		{
			Name:     "Wrong, missing certificates directory",
			Job:      entities.Job{Host: entities.Host{Password: "old"}},
			Expected: entities.Result{Errors: []error{errors.New("keys_directory not specified")}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			sugar := zap.NewExample().Sugar()
			client := mocks.Client{}

			result := InitSecureAPI(context.Background(), sugar, client, &tc.Job)
			if !reflect.DeepEqual(result, tc.Expected) {
				t.Errorf("got:%v, expected:%v", result, tc.Expected)
			}
		})
	}
}
