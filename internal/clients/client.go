package clients

import (
	"context"
	"regexp"
)

// Client interface for all supported clients.
type Client interface {
	GetConfig() Config

	RunCmd(string, *regexp.Regexp) (string, error)
	Connect(ctx context.Context, IP, Port, User, Password string) error
	Close() error
}

// Copier interface for copy files capable clients.
type Copier interface {
	CopyFile(ctx context.Context, local, remote string) (string, error)
}
