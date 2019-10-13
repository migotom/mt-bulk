package mode

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// ChangePassword changes device's admin password.
func ChangePassword(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) ([]entities.CommandResult, error) {
	newPassword, ok := job.Data["new_password"]
	if !ok || newPassword == "" {
		return nil, fmt.Errorf("missing or empty new password for change password operation")
	}

	user, ok := job.Data["user"]
	if !ok || newPassword == "" {
		user = "admin"
	}

	if err := clients.EstablishConnection(ctx, sugar, client, job); err != nil {
		return nil, err
	}
	defer client.Close()

	commands := []entities.Command{
		{Body: fmt.Sprintf("/user/set =numbers=%s =password=%s", user, newPassword), Expect: "!done"},
	}

	results, err := clients.ExecuteCommands(ctx, client, commands)
	if err != nil {
		return results, fmt.Errorf("executing custom commands error %v", err)
	}
	return results, err
}
