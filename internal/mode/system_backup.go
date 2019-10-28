package mode

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

// SystemBackup backups system.
func SystemBackup(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) entities.Result {
	name, ok := job.Data["name"]
	if !ok || name == "" {
		name = "backup"
	}
	name = fmt.Sprintf("%s-%s", name, job.Host.IP)

	backupsStore, ok := job.Data["backups_store"]
	if !ok || backupsStore == "" {
		return entities.Result{Errors: []error{fmt.Errorf("backups_store not specified")}}
	}

	rootDirectory, ok := job.Data["root_directory"]
	if ok && rootDirectory != "" {
		var err error
		backupsStore, err = clients.SecurePathJoin(rootDirectory, backupsStore)
		if err != nil {
			return entities.Result{Errors: []error{err}}
		}
	}

	if _, err := os.Stat(backupsStore); os.IsNotExist(err) {
		if err := os.MkdirAll(backupsStore, os.ModePerm); err != nil {
			return entities.Result{Errors: []error{err}}
		}
	}

	results := make([]entities.CommandResult, 0, 3)

	establishResult, err := clients.EstablishConnection(ctx, sugar, client, job)
	results = append(results, establishResult)
	if err != nil {
		return entities.Result{Errors: []error{err}}
	}
	defer client.Close()

	// prepare sequence of commands to run on device
	commands := []entities.Command{
		{Body: fmt.Sprintf("/system backup save dont-encrypt=yes name=%s", name), SleepMs: 1000},
		{Body: fmt.Sprintf("/export file=%s", name), SleepMs: 5000},
	}

	commandResults, _, err := clients.ExecuteCommands(ctx, client, commands)
	results = append(results, commandResults...)
	if err != nil {
		return entities.Result{Results: results, Errors: []error{fmt.Errorf("executing SystemBackup commands error %v", err)}}
	}

	copier, ok := client.(clients.Copier)
	if !ok {
		return entities.Result{Results: results, Errors: []error{fmt.Errorf("copy file operation not implemented for protocol %v", client)}}
	}

	downloadURLs := make([]string, 0, 2)
	for _, extension := range []string{"backup", "rsc"} {
		var sftpCopyResult entities.CommandResult

		target := filepath.FromSlash(filepath.Join(backupsStore, fmt.Sprintf("%s.%s", name, extension)))
		sftpCopyResult, err = copier.CopyFile(ctx,
			fmt.Sprintf("sftp://%s.%s", name, extension),
			target,
		)

		results = append(results, sftpCopyResult)
		downloadURLs = append(downloadURLs, target)
	}
	if err != nil {
		return entities.Result{Results: results, Errors: []error{err}}
	}

	return entities.Result{Results: results, DownloadURLs: downloadURLs, Errors: []error{err}}
}
