package mode

import (
	"context"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// ChangePassword mode changes device's admin password.
func ChangePassword(ctx context.Context, newService service.NewServiceFunc, config *schema.GeneralConfig, host schema.Host, newPass string) error {
	mt := newService(config, host)

	return mt.HandleSequence(ctx, func(payloadService service.Service) error {
		cmds := []schema.Command{
			{Body: `/user/set =numbers=admin =password=` + newPass, Expect: "!done"},
		}
		return service.ExecuteCommands(ctx, payloadService, cmds)
	})

}
