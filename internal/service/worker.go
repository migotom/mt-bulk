package service

import (
	"context"
	"sort"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/mode"
)

// Worker processing given jobs by jobs channel and sending responses back to results channel.
type Worker struct {
	version         string
	sugar           *zap.SugaredLogger
	jobs            chan entities.Job
	processingHosts []entities.Host
}

// NewWorker returns new worker.
func NewWorker(sugar *zap.SugaredLogger, jobsQueueSize int, version string) *Worker {
	return &Worker{
		sugar:   sugar,
		version: version,
		jobs:    make(chan entities.Job, jobsQueueSize),
	}
}

// ProcessingHost returns true if worker is already processing job for given host.
func (w *Worker) ProcessingHost(host entities.Host) bool {
	return sort.Search(len(w.processingHosts), func(i int) bool { return w.processingHosts[i] == host }) < len(w.processingHosts)
}

// ProcessJobs processes job's channel using given clients configuration.
func (w *Worker) ProcessJobs(ctx context.Context, clientConfig clients.Clients) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-w.jobs:
			if !ok {
				return
			}

			var handler mode.OperationModeFunc
			var client clients.Client

			switch job.Kind {
			case mode.InitPublicKeySSHMode:
				client = clients.NewSSHClient(clientConfig.SSH)
				handler = mode.InitPublicKeySSH
			case mode.InitSecureAPIMode:
				client = clients.NewSSHClient(clientConfig.SSH)
				handler = mode.InitSecureAPI
			case mode.CustomSSHMode:
				client = clients.NewSSHClient(clientConfig.SSH)
				handler = mode.Custom
			case mode.CustomAPIMode:
				client = clients.NewMikrotikAPIClient(clientConfig.MikrotikAPI)
				handler = mode.Custom
			case mode.ChangePasswordMode:
				client = clients.NewMikrotikAPIClient(clientConfig.MikrotikAPI)
				handler = mode.ChangePassword
			case mode.CheckMTbulkVersionMode:
				handler = mode.CheckMTbulkVersion(w.version)
			case mode.SFTPMode:
				client = clients.NewSSHClient(clientConfig.SSH)
				handler = mode.SFTP
			default:
				w.sugar.Infow("unexpected job", "kind", job.Kind)
				continue
			}

			results, err := handler(ctx, w.sugar, client, &job)

			select {
			case <-ctx.Done():
				return
			case job.Result <- entities.Result{Job: job, Results: results, Error: err}:
			}
			close(job.Result)
		}
	}
}
