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
func SystemBackup(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) ([]entities.CommandResult, []string, error) {
	name, ok := job.Data["name"]
	if !ok || name == "" {
		name = "backup"
	}
	name = fmt.Sprintf("%s-%s", name, job.Host.IP)

	backupsStore, ok := job.Data["backups_store"]
	if !ok || backupsStore == "" {
		return nil, nil, fmt.Errorf("backups_store not specified")
	}

	rootDirectory, ok := job.Data["root_directory"]
	if ok && rootDirectory != "" {
		var err error
		backupsStore, err = clients.SecurePathJoin(rootDirectory, backupsStore)
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

	// prepare sequence of commands to run on device
	commands := []entities.Command{
		{Body: fmt.Sprintf("/system backup save dont-encrypt=yes name=%s", name), SleepMs: 1000},
		{Body: fmt.Sprintf("/export file=%s", name), SleepMs: 5000},
	}

	commandResults, err := clients.ExecuteCommands(ctx, client, commands)
	if err != nil {
		err = fmt.Errorf("executing InitPublicKeySSH commands error %v", err)
	}
	results = append(results, commandResults...)

	copier, ok := client.(clients.Copier)
	if !ok {
		return nil, nil, fmt.Errorf("copy file operation not implemented for protocol %v", client)
	}

	downloadURLs := make([]string, 0, 2)
	for _, extension := range []string{"backup", "rsc"} {
		var sftpCopyResult entities.CommandResult

		target := filepath.Join(backupsStore, fmt.Sprintf("%s.%s", name, extension))
		sftpCopyResult, err = copier.CopyFile(ctx,
			fmt.Sprintf("sftp://%s.%s", name, extension),
			target,
		)

		results = append(results, sftpCopyResult)
		downloadURLs = append(downloadURLs, target)
	}
	if err != nil {
		return results, nil, err
	}

	return results, downloadURLs, err
}
