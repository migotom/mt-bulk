package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/migotom/mt-bulk/internal/config"
	"github.com/migotom/mt-bulk/internal/driver"
	"github.com/migotom/mt-bulk/internal/mode"
	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

func configParser(arguments map[string]interface{}, appConfig *schema.GeneralConfig) ([]schema.HostsLoaderFunc, []schema.HostsCleanerFunc, error) {
	var hostsLoaders []schema.HostsLoaderFunc
	var cleaners []schema.HostsCleanerFunc

	appConfig.Service = make(map[string]*schema.Service)

	// config file
	var apiConfigFile string
	apiConfigFile, _ = arguments["-C"].(string)
	if err := config.LoadConfigFile(appConfig, apiConfigFile); err != nil {
		return nil, nil, err
	}

	appConfig.Verbose = !arguments["-s"].(bool)
	appConfig.SkipSummary = arguments["--skip-summary"].(bool)
	appConfig.SkipVersionCheck = arguments["--skip-version-check"].(bool)
	appConfig.IgnoreErrors = !arguments["--exit-on-error"].(bool)

	// TODO
	// add verification of existence in config mikrotik_api and ssh

	// defaults
	if appConfig.VerifySleep == 0 {
		appConfig.VerifySleep = 1000
	}
	if appConfig.Workers == 0 {
		appConfig.Workers = 4
	}
	if appConfig.Certs.Directory == "" {
		appConfig.Certs.Directory = "certs/"
	}

	// parse args
	if workers, ok := arguments["-w"].(string); ok {
		if workers, err := strconv.ParseInt(workers, 10, 64); err == nil {
			appConfig.Workers = int(workers)
		}
	}

	if _, err := os.Stat(appConfig.Certs.Directory); err != nil {
		return nil, nil, fmt.Errorf("missing Certificates Store directory, %s", err)
	}

	if gen, _ := arguments["gen-certs"].(bool); gen {
		if err := service.GenerateCA(appConfig); err != nil {
			return nil, nil, err
		}
		if err := service.GenerateCerts(appConfig, "device"); err != nil {
			return nil, nil, err
		}
		if err := service.GenerateCerts(appConfig, "client"); err != nil {
			return nil, nil, err
		}
		return nil, nil, nil
	}

	if m, _ := arguments["init-secure-api"].(bool); m {
		appConfig.ModeHandler = func(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
			return mode.InitSecureAPIHandler(ctx, service.NewSSHService, config, host)
		}
	}
	if m, _ := arguments["change-password"].(bool); m {
		newPass, ok := arguments["--new"].(string)
		if !ok {
			return nil, nil, fmt.Errorf("missing new password")
		}

		appConfig.ModeHandler = func(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
			return mode.ChangePassword(ctx, service.NewMTAPIService, config, host, newPass)
		}
	}

	if m, _ := arguments["custom-ssh"].(bool); m {
		appConfig.ModeHandler = func(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
			return mode.CustomSSH(ctx, service.NewSSHService, config, host)
		}

		if appConfig.CustomSSHSequence == nil {
			return nil, nil, fmt.Errorf("missing custom-ssh.command sequence in configuration")
		}
		if f, ok := arguments["--commands-file"].(string); ok {
			appConfig.CustomSSHSequence.Command = nil
			if err := config.LoadConfigFile(appConfig.CustomSSHSequence, f); err != nil {
				return nil, nil, err
			}
		}
	}
	if m, _ := arguments["custom-api"].(bool); m {
		appConfig.ModeHandler = func(ctx context.Context, config *schema.GeneralConfig, host schema.Host) error {
			return mode.CustomAPI(ctx, service.NewMTAPIService, config, host)
		}
		if appConfig.CustomAPISequence == nil {
			return nil, nil, fmt.Errorf("missing custom-api.command sequence in configuration")
		}
		if f, ok := arguments["--commands-file"].(string); ok {
			appConfig.CustomAPISequence.Command = nil
			if err := config.LoadConfigFile(appConfig.CustomAPISequence, f); err != nil {
				return nil, nil, err
			}
		}
	}

	if hosts, ok := arguments["<hosts>"].([]string); ok {
		hostsLoaders = append(hostsLoaders, func(parser schema.HostParserFunc) ([]schema.Host, error) {
			return driver.ArgvLoadHosts(parser, hosts)
		})
	}

	if file, ok := arguments["--source-file"].(string); ok {
		hostsLoaders = append(hostsLoaders, func(parser schema.HostParserFunc) ([]schema.Host, error) {
			return driver.FileLoadHosts(parser, file)
		})
	}

	if db, ok := arguments["--source-db"].(bool); ok && db {
		hostsLoaders = append(hostsLoaders, func(parser schema.HostParserFunc) ([]schema.Host, error) {
			return driver.DBSqlLoadHosts(parser, &appConfig.DB)
		})
	}

	return hostsLoaders, cleaners, nil
}
