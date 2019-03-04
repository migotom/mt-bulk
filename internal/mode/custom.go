package mode

import (
	"context"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// CustomAPI mode ...
func CustomAPI(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {

	// use MT API
	mt := service.MTAPI{}
	mt.AppConfig = config
	mt.Host = host

	return mt.HandleSequence(ctx, func(payloadService interface{}) error {
		d := payloadService.(*service.MTAPI)

		// execute commands
		if err := service.ExecuteCommands(ctx, d, config.CustomAPISequence.Command); err != nil {
			return err
		}

		return nil
	})

}

// CustomSSH ...
func CustomSSH(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
	// use MT API
	ssh := service.SSH{}
	ssh.AppConfig = config
	ssh.Host = host

	return ssh.HandleSequence(ctx, func(payloadService interface{}) error {
		d := payloadService.(*service.SSH)

		// execute commands
		if err := service.ExecuteCommands(ctx, d, config.CustomSSHSequence.Command); err != nil {
			return err
		}

		return nil
	})

}
