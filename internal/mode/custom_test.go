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

// func TestCustomAPI(t *testing.T) {
// 	MTAPI := mocks.Service{}

// 	appConfig := schema.GeneralConfig{
// 		CustomAPISequence: &schema.CustomSequence{
// 			Command: []schema.Command{
// 				schema.Command{Body: "/user/print"},
// 			},
// 		},
// 	}
// 	appConfig.Service = make(map[string]*schema.Service)
// 	appConfig.Service["mikrotik_api"] = &schema.Service{}

// 	if err := CustomAPI(context.Background(), MTAPI.GetService, &appConfig, entities.Host{}); err != nil {
// 		t.Errorf("not expected error %v", err)
// 	}
// 	if !reflect.DeepEqual(MTAPI.CommandsExecuted, []string{"/user/print"}) {
// 		t.Errorf("not expected commands %v", MTAPI.CommandsExecuted)
// 	}
// }

// func TestCustomSSH(t *testing.T) {
// 	SSHAPI := mocks.Service{}

// 	appConfig := schema.GeneralConfig{
// 		CustomSSHSequence: &schema.CustomSequence{
// 			Command: []schema.Command{
// 				schema.Command{Body: "/user print"},
// 			},
// 		},
// 	}
// 	appConfig.Service = make(map[string]*schema.Service)
// 	appConfig.Service["ssh"] = &schema.Service{}

// 	if err := CustomSSH(context.Background(), SSHAPI.GetService, &appConfig, entities.Host{}); err != nil {
// 		t.Errorf("not expected error %v", err)
// 	}
// 	if !reflect.DeepEqual(SSHAPI.CommandsExecuted, []string{"/user print"}) {
// 		t.Errorf("not expected commands %v", SSHAPI.CommandsExecuted)
// 	}
// }
