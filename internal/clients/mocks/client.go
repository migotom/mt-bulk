package mocks

import (
	"context"
	"fmt"
	"regexp"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// func NewSSHClient(config *schema.GeneralConfig, host entities.Host) interface{} {
// 	return SSHClient{}
// }

// Client mocked.
type Client struct {
	Host entities.Host
}

// func (t *Client) GetClient(config *schema.GeneralConfig, host entities.Host) client.Client {
// 	t.AppConfig = config
// 	t.Host = host
// 	return t
// }

func (t Client) GetConfig() clients.Config {
	return clients.Config{Retries: 1}
}

func (t Client) CopyFile(ctx context.Context, local, remote string) (string, error) {
	return fmt.Sprintf("/<mt-bulk>copy sftp://%s %s", local, remote), nil
}

func (t Client) Connect(ctx context.Context, IP, Port, User, Password string) (err error) {
	return nil
}

func (t Client) RunCmd(val string, re *regexp.Regexp) (string, error) {
	return val, nil
}

func (t Client) Close() error {
	return nil
}
