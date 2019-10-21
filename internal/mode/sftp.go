package mode

import (
	"context"
	"fmt"
	"strings"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

// SFTP initializes SSH public key authentication.
func SFTP(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) ([]entities.CommandResult, []string, error) {
	source, ok := job.Data["source"]
	if !ok || source == "" {
		return nil, nil, fmt.Errorf("source not specified")
	}

	target, ok := job.Data["target"]
	if !ok || target == "" {
		return nil, nil, fmt.Errorf("target not specified")
	}

	rootDirectory, ok := job.Data["root_directory"]
	if ok && rootDirectory != "" {
		var err error
		source, err = clients.SecurePathJoin(rootDirectory, source)
		if err != nil {
			return nil, nil, err
		}
		target, err = clients.SecurePathJoin(rootDirectory, target)
		if err != nil {
			return nil, nil, err
		}
	}

	hasSFTPside := false
	for _, location := range []string{source, target} {
		if strings.Index(location, "sftp://") == 0 {
			hasSFTPside = true
			break
		}
	}
	if !hasSFTPside {
		return nil, nil, fmt.Errorf("at least one side of sftp transfer has to be remote, syntax like: sftp://remote_file_name.txt")

	}

	results := make([]entities.CommandResult, 0, 2)

	establishResult, err := clients.EstablishConnection(ctx, sugar, client, job)
	results = append(results, establishResult)
	if err != nil {
		return results, nil, err
	}
	defer client.Close()

	copier, ok := client.(clients.Copier)
	if !ok {
		return results, nil, fmt.Errorf("copy file operation not implemented for protocol %v", client)
	}

	var sftpCopyResult entities.CommandResult
	sftpCopyResult, err = copier.CopyFile(ctx, source, target)
	results = append(results, sftpCopyResult)
	if err != nil {
		return results, nil, err
	}

	if strings.Index(source, "sftp://") == 0 && strings.Index(target, "sftp://") == -1 {
		return results, []string{target}, err
	}
	return results, nil, err
}
