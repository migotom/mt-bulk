package mode

import (
	"context"
	"fmt"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// Custom executes by client custom job.
func Custom(ctx context.Context, client clients.Client, job *entities.Job) ([]entities.CommandResult, error) {
	if err := EstablishConnection(ctx, client, &job.Host); err != nil {
		return nil, err
	}
	defer client.Close()

	results, err := ExecuteCommands(ctx, client, job.Commands)
	if err != nil {
		return results, fmt.Errorf("executing custom commands error %v", err)
	}
	return results, err
}
