package mocks

import (
	"context"
	"fmt"
	"regexp"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// Client mocked.
type Client struct {
	Host entities.Host
}

// GetConfig mock.
func (t Client) GetConfig() clients.Config {
	return clients.Config{Retries: 1}
}

// CopyFile mock.
func (t Client) CopyFile(ctx context.Context, local, remote string) (entities.CommandResult, error) {
	return entities.CommandResult{Body: fmt.Sprintf("/<mt-bulk>copy sftp://%s %s", local, remote)}, nil
}

// Connect mock.
func (t Client) Connect(ctx context.Context, IP, Port, User, Password string) (err error) {
	return nil
}

// RunCmd mock.
func (t Client) RunCmd(val string, re *regexp.Regexp) (string, error) {
	return val, nil
}

// Close mock.
func (t Client) Close() {
}
