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
func InitPublicKeySSH(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) ([]entities.CommandResult, error) {
	certificatesDirectory, ok := job.Data["keys_directory"]
	if !ok || certificatesDirectory == "" {
		return nil, fmt.Errorf("keys_directory not specified")
	}

	results := make([]entities.CommandResult, 0, 3)

	establishResult, err := clients.EstablishConnection(ctx, sugar, client, job)
	results = append(results, establishResult)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	copier, ok := client.(clients.Copier)
	if !ok {
		return nil, fmt.Errorf("copy file operation not implemented for protocol %v", client)
	}

	var sftpCopyResult string
	sftpCopyResult, err = copier.CopyFile(ctx, filepath.Join(certificatesDirectory, "id_rsa.pub"), "id_rsa.pub")
	results = append(results, entities.CommandResult{Body: sftpCopyResult, Error: err})
	if err != nil {
		return results, fmt.Errorf("file copy error %v", err)
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
	return results, err
}
