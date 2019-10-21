package mode

import (
	"context"
	"fmt"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

// Custom executes by client custom job.
func Custom(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) ([]entities.CommandResult, []string, error) {
	results := make([]entities.CommandResult, 0, 8)

	establishResult, err := clients.EstablishConnection(ctx, sugar, client, job)
	results = append(results, establishResult)
	if err != nil {
		return results, nil, err
	}
	defer client.Close()

	commandResults, err := clients.ExecuteCommands(ctx, client, job.Commands)
	results = append(results, commandResults...)
	if err != nil {
		return results, nil, fmt.Errorf("executing custom commands error %v", err)
	}
	return results, nil, err
}
