package mode

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// InitPublicKeySSH initializes SSH public key authentication.
func InitPublicKeySSH(ctx context.Context, client clients.Client, job *entities.Job) (results []string, err error) {
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

	var sftpCopyResult string
	sftpCopyResult, err = copier.CopyFile(ctx, filepath.Join(certificatesDirectory, "id_rsa.pub"), "id_rsa.pub")
	if err != nil {
		return nil, fmt.Errorf("file copy error %v", err)
	}
	results = append(results, sftpCopyResult)

	// prepare sequence of commands to run on device
	commands := []entities.Command{
		{Body: `/user ssh-keys import public-key-file=id_rsa.pub`},
	}

	sshResults, err := ExecuteCommands(ctx, client, commands)
	if err != nil {
		err = fmt.Errorf("executing InitPublicKeySSH commands error %v", err)
	}
	results = append(results, sshResults...)
	results = append(results, "/// HINT: remember, enabling ssh-key on routerOS disables authentication by password.")
	return results, err
}
