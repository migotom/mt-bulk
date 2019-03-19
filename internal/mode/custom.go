package mode

import (
	"context"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// CustomAPI executes custom sequence of commands using Mikrotik SSL API.
func CustomAPI(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
	mt := config.Service["mikrotik_api"].Interface.(service.Service)
	mt.SetConfig(config)
	mt.SetHost(host)

	return mt.HandleSequence(ctx, func(payloadService service.Service) error {
		return service.ExecuteCommands(ctx, payloadService, config.CustomAPISequence.Command)
	})

}

// CustomSSH executes custom sequence of commands using SSH protocol.
func CustomSSH(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
	ssh := config.Service["ssh"].Interface.(service.Service)
	ssh.SetConfig(config)
	ssh.SetHost(host)

	return ssh.HandleSequence(ctx, func(payloadService service.Service) error {
		return service.ExecuteCommands(ctx, payloadService, config.CustomSSHSequence.Command)
	})

}
