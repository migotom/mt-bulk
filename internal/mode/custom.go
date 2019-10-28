package mode

import (
	"context"
	"fmt"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

// Custom executes by client custom job.
func Custom(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) entities.Result {
	results := make([]entities.CommandResult, 0, 8)

	establishResult, err := clients.EstablishConnection(ctx, sugar, client, job)
	results = append(results, establishResult)
	if err != nil {
		return entities.Result{Results: results, Errors: []error{err}}
	}
	defer client.Close()

	commandResults, _, err := clients.ExecuteCommands(ctx, client, job.Commands)
	results = append(results, commandResults...)
	if err != nil {
		return entities.Result{Results: results, Errors: []error{fmt.Errorf("executing custom commands error %v", err)}}
	}
	return entities.Result{Results: results}
}
