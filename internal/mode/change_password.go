package mode

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// ChangePassword changes device's admin password.
func ChangePassword(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) entities.Result {
	newPassword, ok := job.Data["new_password"]
	if !ok || newPassword == "" {
		return entities.Result{Errors: []error{fmt.Errorf("missing or empty new password for change password operation")}}
	}

	user, ok := job.Data["user"]
	if !ok || newPassword == "" {
		user = "admin"
	}

	results := make([]entities.CommandResult, 0, 2)

	establishResult, err := clients.EstablishConnection(ctx, sugar, client, job)
	results = append(results, establishResult)
	if err != nil {
		return entities.Result{Results: results, Errors: []error{err}}
	}
	defer client.Close()

	commands := []entities.Command{
		{Body: fmt.Sprintf("/user/set =numbers=%s =password=%s", user, newPassword), Expect: "!done"},
	}

	commandResults, _, err := clients.ExecuteCommands(ctx, client, commands)
	results = append(results, commandResults...)
	if err != nil {
		return entities.Result{Results: results, Errors: []error{fmt.Errorf("executing custom commands error %v", err)}}
	}
	return entities.Result{Results: results}
}
