package mode

import (
	"context"
	"fmt"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

// Custom executes by client custom job.
func Custom(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) ([]entities.CommandResult, error) {
	if err := clients.EstablishConnection(ctx, sugar, client, job); err != nil {
		return nil, err
	}
	defer client.Close()

	results, err := clients.ExecuteCommands(ctx, client, job.Commands)
	if err != nil {
		return results, fmt.Errorf("executing custom commands error %v", err)
	}
	return results, err
}
