package mode

import (
	"context"
	"fmt"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

// SFTP initializes SSH public key authentication.
func SFTP(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) ([]entities.CommandResult, error) {
	source, ok := job.Data["source"]
	if !ok || source == "" {
		return nil, fmt.Errorf("source not specified")
	}

	target, ok := job.Data["target"]
	if !ok || target == "" {
		return nil, fmt.Errorf("target not specified")
	}

	results := make([]entities.CommandResult, 0, 2)

	establishResult, err := clients.EstablishConnection(ctx, sugar, client, job)
	results = append(results, establishResult)
	if err != nil {
		return results, err
	}
	defer client.Close()

	copier, ok := client.(clients.Copier)
	if !ok {
		return results, fmt.Errorf("copy file operation not implemented for protocol %v", client)
	}

	var sftpCopyResult entities.CommandResult
	sftpCopyResult, err = copier.CopyFile(ctx, source, target)
	results = append(results, sftpCopyResult)
	if err != nil {
		return results, err
	}

	return results, err
}
