package mode

import (
	"context"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// ChangePassword mode ...
func ChangePassword(ctx context.Context, config *schema.GeneralConfig, host schema.Host, newPass string) error {

	// use MT API
	mt := service.MTAPI{}
	mt.AppConfig = config
	mt.Host = host

	return mt.HandleSequence(ctx, func(payloadService interface{}) error {
		d := payloadService.(*service.MTAPI)

		// prepare sequence of commands to run on device
		cmds := []schema.Command{
			{Body: `/user/set =numbers=admin =password=` + newPass, Expect: "!done"},
		}

		// execute commands
		if err := service.ExecuteCommands(ctx, d, cmds); err != nil {
			return err
		}

		return nil
	})

}
