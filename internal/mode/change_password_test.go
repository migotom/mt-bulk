package mode

import (
	"context"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service/mocks"
)

func TestChangePassword(t *testing.T) {
	MTAPI := mocks.Service{}
	appConfig := schema.GeneralConfig{}
	appConfig.Service = make(map[string]*schema.Service)
	appConfig.Service["mikrotik_api"] = &schema.Service{}

	if err := ChangePassword(context.Background(), MTAPI.GetService, &appConfig, schema.Host{}, "new_pass"); err != nil {
		t.Errorf("not expected error %v", err)
	}
	if !reflect.DeepEqual(MTAPI.CommandsExecuted, []string{"/user/set =numbers=admin =password=new_pass"}) {
		t.Errorf("not expected commands %v", MTAPI.CommandsExecuted)
	}
}
