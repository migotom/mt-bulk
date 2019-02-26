package mode

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// InitSecureAPIHandler mode initializes device for secure API usage (copies and sets up certificate)
func InitSecureAPIHandler(config *schema.GeneralConfig, host schema.Host) error {

	// use SSH to initialize secure API
	ssh := service.SSH{}
	ssh.AppConfig = config
	ssh.Host = host
	log.Printf("[InitSecureAPI] IP %s", host.IP)

	return ssh.HandleSequence(func(ctx interface{}) error {
		d := ctx.(*service.SSH)

		// copy certificate to device
		if err := d.CopyFile(filepath.Join(config.Certs.Directory, "device.crt"), "mtbulkdevice.crt"); err != nil {
			return fmt.Errorf("file copy error %v", err)
		}
		if err := d.CopyFile(filepath.Join(config.Certs.Directory, "device.key"), "mtbulkdevice.key"); err != nil {
			return fmt.Errorf("file copy error %v", err)
		}

		// prepare sequence of commands to run on device
		cmds := []schema.Command{
			{Body: `/ip service set api-ssl certificate=none`},
			{Body: `/certificate print detail`, MatchPrefix: "c", Match: `(?m)^\s+(\d+).*mtbulkdevice`},
			{Body: `/certificate remove %{c1}`},
			//{Body: `/certificate import file-name=mtbulkdevice.crt passphrase=""`, Expect: "certificates-imported: 1"},
			//{Body: `/certificate import file-name=mtbulkdevice.key passphrase=""`, Expect: "private-keys-imported: 1", Sleep: schema.Duration{Duration: time.Duration(config.VerifySleep) * time.Millisecond}},
			//{Body: `/ip service set api-ssl certificate=mtbulkdevice.crt`},
		}

		// execute commands
		if err := service.ExecuteCommands(ssh, cmds); err != nil {
			return fmt.Errorf("executing command error %v", err)
		}

		return nil
	})

}
