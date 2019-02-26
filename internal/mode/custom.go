package mode

import (
	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// CustomAPI mode ...
func CustomAPI(config *schema.GeneralConfig, host schema.Host) error {

	// use MT API
	mt := service.MTAPI{}
	mt.AppConfig = config
	mt.Host = host

	return mt.HandleSequence(func(ctx interface{}) error {
		d := ctx.(*service.MTAPI)

		// execute commands
		if err := service.ExecuteCommands(d, config.CustomAPISequence.Command); err != nil {
			return err
		}

		return nil
	})

}

// CustomSSH ...
func CustomSSH(config *schema.GeneralConfig, host schema.Host) error {
	// use MT API
	ssh := service.SSH{}
	ssh.AppConfig = config
	ssh.Host = host

	return ssh.HandleSequence(func(ctx interface{}) error {
		d := ctx.(*service.SSH)

		// execute commands
		if err := service.ExecuteCommands(d, config.CustomSSHSequence.Command); err != nil {
			return err
		}

		return nil
	})

}
