package mode

import (
	"context"
	"reflect"
	"testing"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service/mocks"
)

func TestInitSecureAPIHandler(t *testing.T) {
	MTAPI := mocks.Service{}

	appConfig := schema.GeneralConfig{}
	appConfig.Service = make(map[string]*schema.Service)
	appConfig.Service["ssh"] = &schema.Service{}

	if err := InitSecureAPIHandler(context.Background(), MTAPI.GetService, &appConfig, schema.Host{}); err != nil {
		t.Errorf("not expected error %v", err)
	}
	if !reflect.DeepEqual(MTAPI.CommandsExecuted, []string{
		`/ip service set api-ssl certificate=none`,
		`/certificate print detail`,
		`/certificate remove %{c1}`,
		`/certificate import file-name=mtbulkdevice.crt passphrase=""`,
		`/certificate import file-name=mtbulkdevice.key passphrase=""`,
		`/ip service set api-ssl disabled=no certificate=mtbulkdevice.crt`,
	}) {
		t.Errorf("not expected commands %v", MTAPI.CommandsExecuted)
	}
	if !reflect.DeepEqual(MTAPI.FilesCopied, []string{"mtbulkdevice.crt", "mtbulkdevice.key"}) {
		t.Errorf("not expected copied files %v", MTAPI.FilesCopied)
	}
}
