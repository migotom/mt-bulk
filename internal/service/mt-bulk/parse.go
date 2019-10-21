package mtbulk

import (
	"context"
	"errors"
	"fmt"

	"github.com/migotom/mt-bulk/internal/service"

	"github.com/migotom/mt-bulk/internal/config"
	"github.com/migotom/mt-bulk/internal/mode"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/driver"
	"github.com/migotom/mt-bulk/internal/entities"
)

func configParser(arguments map[string]interface{}, version string) (mtbulkConfig Config, jobsLoaders []entities.JobsLoaderFunc, jobTemplate entities.Job, err error) {

	mtbulkConfig = Config{}
	mtbulkConfig.Service = service.NewConfig(version)
	mtbulkConfig.Service.Clients.SSH = clients.NewConfig(clients.SSHDefaultPort)
	mtbulkConfig.Service.Clients.MikrotikAPI = clients.NewConfig(clients.MikrotikAPIDefaultPort)

	configFileName, _ := arguments["-C"].(string)
	if err := config.LoadConfigFile(&mtbulkConfig, configFileName); err != nil {
		return Config{}, nil, entities.Job{}, err
	}

	if mtbulkConfig.Version < 2 {
		return Config{}, nil, entities.Job{}, errors.New("incompatible configuration version, required version 2 or above")
	}
	if gen, _ := arguments["gen-api-certs"].(bool); gen {
		if err := clients.GenerateCA(mtbulkConfig.Service.Clients.MikrotikAPI.KeyStore); err != nil {
			return Config{}, nil, entities.Job{}, err
		}
		if err := clients.GenerateCerts(mtbulkConfig.Service.Clients.MikrotikAPI.KeyStore, "device"); err != nil {
			return Config{}, nil, entities.Job{}, err
		}
		if err := clients.GenerateCerts(mtbulkConfig.Service.Clients.MikrotikAPI.KeyStore, "client"); err != nil {
			return Config{}, nil, entities.Job{}, err
		}
		return Config{}, nil, entities.Job{}, nil
	}
	if gen, _ := arguments["gen-ssh-keys"].(bool); gen {
		if err := clients.GenerateKeys(mtbulkConfig.Service.Clients.SSH.KeyStore); err != nil {
			return Config{}, nil, entities.Job{}, err
		}
		return Config{}, nil, entities.Job{}, nil
	}

	if m, _ := arguments["init-secure-api"].(bool); m {
		jobTemplate = entities.Job{
			Kind: mode.InitSecureAPIMode,
			Data: map[string]string{"keys_directory": mtbulkConfig.Service.Clients.MikrotikAPI.KeyStore},
		}
	}

	if m, _ := arguments["init-publickey-ssh"].(bool); m {
		jobTemplate = entities.Job{
			Kind: mode.InitPublicKeySSHMode,
			Data: map[string]string{"keys_directory": mtbulkConfig.Service.Clients.SSH.KeyStore},
		}
	}

	if m, _ := arguments["change-password"].(bool); m {
		newPass, ok := arguments["--new"].(string)
		if !ok {
			return Config{}, nil, entities.Job{}, fmt.Errorf("missing new password")
		}

		jobTemplate = entities.Job{
			Kind: mode.ChangePasswordMode,
			Data: map[string]string{"new_password": newPass},
		}

		user, ok := arguments["--user"].(string)
		if ok {
			jobTemplate.Data["user"] = user
		}
	}

	if m, _ := arguments["custom-ssh"].(bool); m {
		if mtbulkConfig.CustomSSHSequence == nil {
			return Config{}, nil, entities.Job{}, fmt.Errorf("missing custom-ssh.command sequence in configuration")
		}
		if f, ok := arguments["--commands-file"].(string); ok {
			mtbulkConfig.CustomSSHSequence.Command = nil
			if err := config.LoadConfigFile(mtbulkConfig.CustomSSHSequence, f); err != nil {
				return Config{}, nil, entities.Job{}, err
			}
		}
		jobTemplate = entities.Job{
			Kind:     mode.CustomSSHMode,
			Commands: mtbulkConfig.CustomSSHSequence.Command,
		}
	}
	if m, _ := arguments["custom-api"].(bool); m {
		if mtbulkConfig.CustomAPISequence == nil {
			return Config{}, nil, entities.Job{}, fmt.Errorf("missing custom-api.command sequence in configuration")
		}
		if f, ok := arguments["--commands-file"].(string); ok {
			mtbulkConfig.CustomAPISequence.Command = nil
			if err := config.LoadConfigFile(mtbulkConfig.CustomAPISequence, f); err != nil {
				return Config{}, nil, entities.Job{}, err
			}
		}
		jobTemplate = entities.Job{
			Kind:     mode.CustomAPIMode,
			Commands: mtbulkConfig.CustomAPISequence.Command,
		}

	}

	if m, _ := arguments["sftp"].(bool); m {
		source, ok := arguments["<source>"].(string)
		if !ok {
			return Config{}, nil, entities.Job{}, fmt.Errorf("missing source")
		}

		target, ok := arguments["<target>"].(string)
		if !ok {
			return Config{}, nil, entities.Job{}, fmt.Errorf("missing source")
		}

		jobTemplate = entities.Job{
			Kind: mode.SFTPMode,
			Data: map[string]string{"source": source, "target": target},
		}
	}

	if m, _ := arguments["system-backup"].(bool); m {
		name, ok := arguments["--name"].(string)
		if !ok {
			return Config{}, nil, entities.Job{}, fmt.Errorf("missing backup name")
		}

		backupsStore, ok := arguments["--backup-store"].(string)
		if !ok {
			return Config{}, nil, entities.Job{}, fmt.Errorf("missing local backup store location")
		}

		jobTemplate = entities.Job{
			Kind: mode.SystemBackupMode,
			Data: map[string]string{"name": name, "backups_store": backupsStore},
		}
	}

	if hosts, ok := arguments["<hosts>"].([]string); ok {
		jobsLoaders = append(jobsLoaders, func(ctx context.Context, jobTemplate entities.Job) ([]entities.Job, error) {
			return driver.ArgvLoadJobs(ctx, jobTemplate, hosts)
		})
	}

	if file, ok := arguments["--source-file"].(string); ok {
		jobsLoaders = append(jobsLoaders, func(ctx context.Context, jobTemplate entities.Job) ([]entities.Job, error) {
			return driver.FileLoadJobs(ctx, jobTemplate, file)
		})
	}

	if db, ok := arguments["--source-db"].(bool); ok && db {
		jobsLoaders = append(jobsLoaders, func(ctx context.Context, jobTemplate entities.Job) ([]entities.Job, error) {
			return driver.DBSqlLoadJobs(ctx, jobTemplate, &mtbulkConfig.DB)
		})
	}

	return mtbulkConfig, jobsLoaders, jobTemplate, nil
}
