package mode

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

// InitPublicKeySSH initializes SSH public key authentication.
func InitPublicKeySSH(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) ([]entities.CommandResult, []string, error) {
	certificatesDirectory, ok := job.Data["keys_directory"]
	if !ok || certificatesDirectory == "" {
		return nil, nil, fmt.Errorf("keys_directory not specified")
	}

	rootDirectory, ok := job.Data["root_directory"]
	if ok && rootDirectory != "" {
		var err error
		certificatesDirectory, err = clients.SecurePathJoin(rootDirectory, certificatesDirectory)
		if err != nil {
			return nil, nil, err
		}
	}

	results := make([]entities.CommandResult, 0, 3)

	establishResult, err := clients.EstablishConnection(ctx, sugar, client, job)
	results = append(results, establishResult)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()

	copier, ok := client.(clients.Copier)
	if !ok {
		return nil, nil, fmt.Errorf("copy file operation not implemented for protocol %v", client)
	}

	var sftpCopyResult entities.CommandResult
	sftpCopyResult, err = copier.CopyFile(ctx, filepath.Join(certificatesDirectory, "id_rsa.pub"), "sftp://id_rsa.pub")
	results = append(results, sftpCopyResult)
	if err != nil {
		return results, nil, err
	}

	// prepare sequence of commands to run on device
	commands := []entities.Command{
		{Body: `/user ssh-keys import public-key-file=id_rsa.pub`},
	}

	commandResults, err := clients.ExecuteCommands(ctx, client, commands)
	if err != nil {
		err = fmt.Errorf("executing InitPublicKeySSH commands error %v", err)
	}
	results = append(results, commandResults...)
	return results, nil, err
}
