package mode

import (
	"context"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// CustomAPI executes custom sequence of commands using Mikrotik SSL API.
func CustomAPI(ctx context.Context, newService service.NewServiceFunc, config *schema.GeneralConfig, host schema.Host) error {
	mt := newService(config, host)

	return mt.HandleSequence(ctx, func(payloadService service.Service) error {
		return service.ExecuteCommands(ctx, payloadService, config.CustomAPISequence.Command)
	})

}

// CustomSSH executes custom sequence of commands using SSH protocol.
func CustomSSH(ctx context.Context, newService service.NewServiceFunc, config *schema.GeneralConfig, host schema.Host) error {
	ssh := newService(config, host)

	return ssh.HandleSequence(ctx, func(payloadService service.Service) error {
		return service.ExecuteCommands(ctx, payloadService, config.CustomSSHSequence.Command)
	})

}
