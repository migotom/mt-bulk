package mode

import (
	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// ChangePassword mode ...
func ChangePassword(config *schema.GeneralConfig, host schema.Host, newPass string) error {

	// use MT API
	mt := service.MTAPI{}
	mt.AppConfig = config
	mt.Host = host

	return mt.HandleSequence(func(ctx interface{}) error {
		d := ctx.(*service.MTAPI)

		// prepare sequence of commands to run on device
		cmds := []schema.Command{
			{Body: `/user/set =numbers=admin =password=` + newPass, Expect: "!done"},
		}

		// execute commands
		if err := service.ExecuteCommands(d, cmds); err != nil {
			return err
		}

		return nil
	})

}
