package mocks

import (
	"context"
	"regexp"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// Service mocked.
type Service struct {
	CommandsExecuted []string
	FilesCopied      []string

	AppConfig *schema.GeneralConfig
	Host      schema.Host
}

func (t *Service) CopyFile(ctx context.Context, local, remote string) error {
	t.FilesCopied = append(t.FilesCopied, remote)
	return nil
}

func (t *Service) GetUser() string {
	return "testusert"
}
func (t *Service) GetPasswords() []string {
	return []string{"1stpass", "2ndpass"}
}

func (t *Service) GetPort() string {
	return "22"
}
func (t *Service) GetDevice() *service.GenericDevice {
	return &service.GenericDevice{AppConfig: t.AppConfig}
}

func (t *Service) SetHost(host schema.Host) {
	t.Host = host
}
func (t *Service) SetConfig(config *schema.GeneralConfig) {
	t.AppConfig = config
}

func (t *Service) HandleSequence(ctx context.Context, handler service.HandlerFunc) error {
	return handler(t)
}
func (t *Service) RunCmd(val string, re *regexp.Regexp) (string, error) {
	t.CommandsExecuted = append(t.CommandsExecuted, val)
	return "", nil
}

func (t *Service) Close() error {
	return nil
}
