package mode

import (
	"context"
	"fmt"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// ChangePassword changes device's admin password.
func ChangePassword(ctx context.Context, client clients.Client, job *entities.Job) ([]string, error) {
	newPassword, ok := job.Data["new_password"]
	if !ok || newPassword == "" {
		return nil, fmt.Errorf("missing or empty new password for change password operation")
	}

	user, ok := job.Data["user"]
	if !ok || newPassword == "" {
		user = "admin"
	}

	if err := EstablishConnection(ctx, client, &job.Host); err != nil {
		return nil, err
	}
	defer client.Close()

	commands := []entities.Command{
		{Body: fmt.Sprintf("/user/set =numbers=%s =password=%s", user, newPassword), Expect: "!done"},
	}

	results, err := ExecuteCommands(ctx, client, commands)
	if err != nil {
		return nil, fmt.Errorf("executing custom commands error %v", err)
	}
	return results, nil
}
