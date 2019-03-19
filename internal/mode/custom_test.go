package mode

import (
	"context"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service/mocks"
)

func TestCustomAPI(t *testing.T) {
	MTAPI := mocks.Service{}

	appConfig := schema.GeneralConfig{
		CustomAPISequence: schema.CustomSequence{
			Command: []schema.Command{
				schema.Command{Body: "/user/print"},
			},
		},
	}
	appConfig.Service = make(map[string]*schema.Service)
	appConfig.Service["mikrotik_api"] = &schema.Service{
		Interface: &MTAPI,
	}

	if err := CustomAPI(context.Background(), &appConfig, schema.Host{}); err != nil {
		t.Errorf("not expected error %v", err)
	}
	if !reflect.DeepEqual(MTAPI.CommandsExecuted, []string{"/user/print"}) {
		t.Errorf("not expected commands %v", MTAPI.CommandsExecuted)
	}
}

func TestCustomSSH(t *testing.T) {
	SSHAPI := mocks.Service{}

	appConfig := schema.GeneralConfig{
		CustomSSHSequence: schema.CustomSequence{
			Command: []schema.Command{
				schema.Command{Body: "/user print"},
			},
		},
	}
	appConfig.Service = make(map[string]*schema.Service)
	appConfig.Service["ssh"] = &schema.Service{
		Interface: &SSHAPI,
	}

	if err := CustomSSH(context.Background(), &appConfig, schema.Host{}); err != nil {
		t.Errorf("not expected error %v", err)
	}
	if !reflect.DeepEqual(SSHAPI.CommandsExecuted, []string{"/user print"}) {
		t.Errorf("not expected commands %v", SSHAPI.CommandsExecuted)
	}
}
