package mode

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// InitSecureAPI initializes Mikrotik secure API using SSH client.
func InitSecureAPI(ctx context.Context, client clients.Client, job *entities.Job) (results []entities.CommandResult, err error) {
	certificatesDirectory, ok := job.Data["keys_directory"]
	if !ok || certificatesDirectory == "" {
		return nil, fmt.Errorf("keys_directory not specified")
	}

	if err := EstablishConnection(ctx, client, &job.Host); err != nil {
		return nil, err
	}
	defer client.Close()

	copier, ok := client.(clients.Copier)
	if !ok {
		return nil, fmt.Errorf("copy file operation not implemented for protocol %v", client)
	}

	config := client.GetConfig()

	var sftpCopyResult string
	sftpCopyResult, err = copier.CopyFile(ctx, filepath.Join(certificatesDirectory, "device.crt"), "mtbulkdevice.crt")
	results = append(results, entities.CommandResult{Body: sftpCopyResult, Error: err})
	if err != nil {
		return results, fmt.Errorf("file copy error %v", err)
	}

	sftpCopyResult, err = copier.CopyFile(ctx, filepath.Join(certificatesDirectory, "device.key"), "mtbulkdevice.key")
	results = append(results, entities.CommandResult{Body: sftpCopyResult, Error: err})
	if err != nil {
		return results, fmt.Errorf("file copy error %v", err)
	}

	// prepare sequence of commands to run on device
	commands := []entities.Command{
		{Body: `/ip service set api-ssl certificate=none`},
		{Body: `/certificate print detail`, MatchPrefix: "c", Match: `(?m)^\s+(\d+).*mtbulkdevice`},
		{Body: `/certificate remove %{c1}`},
		{Body: `/certificate import file-name=mtbulkdevice.crt passphrase=""`, Expect: "certificates-imported: 1"},
		{Body: `/certificate import file-name=mtbulkdevice.key passphrase=""`, Expect: "private-keys-imported: 1", SleepMs: config.VerifySleepMs},
		{Body: `/ip service set api-ssl disabled=no certificate=mtbulkdevice.crt`},
	}

	sshResults, err := ExecuteCommands(ctx, client, commands)
	if err != nil {
		err = fmt.Errorf("executing InitSecureAPI commands error %v", err)
	}
	results = append(results, sshResults...)

	return results, err
}
