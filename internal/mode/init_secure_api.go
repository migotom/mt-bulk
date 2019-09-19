package mode

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// InitSecureAPIHandler mode initializes device for secure API usage (copies and sets up certificate)
func InitSecureAPIHandler(ctx context.Context, newService service.NewServiceFunc, config *schema.GeneralConfig, host schema.Host) error {
	ssh := newService(config, host)
	log.Printf("[InitSecureAPI] IP %s", host.IP)

	return ssh.HandleSequence(ctx, func(payloadService service.Service) error {
		d := payloadService.(service.CopyFiler)

		// copy certificate to device
		if err := d.CopyFile(ctx, filepath.Join(config.Certs.Directory, "device.crt"), "mtbulkdevice.crt"); err != nil {
			return fmt.Errorf("file copy error %v", err)
		}
		if err := d.CopyFile(ctx, filepath.Join(config.Certs.Directory, "device.key"), "mtbulkdevice.key"); err != nil {
			return fmt.Errorf("file copy error %v", err)
		}

		// prepare sequence of commands to run on device
		cmds := []schema.Command{
			{Body: `/ip service set api-ssl certificate=none`},
			{Body: `/certificate print detail`, MatchPrefix: "c", Match: `(?m)^\s+(\d+).*mtbulkdevice`},
			{Body: `/certificate remove %{c1}`},
			{Body: `/certificate import file-name=mtbulkdevice.crt passphrase=""`, Expect: "certificates-imported: 1"},
			{Body: `/certificate import file-name=mtbulkdevice.key passphrase=""`, Expect: "private-keys-imported: 1", Sleep: schema.Duration{Duration: time.Duration(config.VerifySleep) * time.Millisecond}},
			{Body: `/ip service set api-ssl disabled=no certificate=mtbulkdevice.crt`},
		}

		if err := service.ExecuteCommands(ctx, payloadService, cmds); err != nil {
			return fmt.Errorf("executing command error %v", err)
		}

		return nil
	})

}
