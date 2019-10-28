package service

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/kvdb"
	"github.com/migotom/mt-bulk/internal/vulnerabilities"
)

// Service of set of workers processing jobs on provided Mikrotik devices.
type Service struct {
	sugar  *zap.SugaredLogger
	config Config
	kv     kvdb.KV
	Jobs   chan entities.Job

	VulnerabilitiesManager *vulnerabilities.Manager
}

// NewService returns new service.
func NewService(sugar *zap.SugaredLogger, kv kvdb.KV, config Config) *Service {
	return &Service{
		sugar:                  sugar,
		config:                 config,
		kv:                     kv,
		Jobs:                   make(chan entities.Job, config.Workers),
		VulnerabilitiesManager: vulnerabilities.NewManager(sugar, config.CVEURL, kv),
	}
}

// Listen runs worker pool of SSH/Mikrotik SSL API clients processing jobs channel and sending jobs execution result into separate channel.
func (service *Service) Listen(ctx context.Context, cancel context.CancelFunc) {
	wg := new(sync.WaitGroup)

	workerPool := NewWorkerPool(service.config.Workers)
	for i := 0; i < service.config.Workers; i++ {
		w := NewWorker(service.sugar, 8, service.config.Version, service.kv, service.VulnerabilitiesManager)
		workerPool.Add(w)

		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()

			w.ProcessJobs(ctx, service.config.Clients)
		}(w)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		service.VulnerabilitiesManager.Listen(ctx)
	}()

	wg.Add(1)
	go func() {
		defer cancel()
		defer workerPool.Close()
		defer wg.Done()

		// jobs dispatcher
		for {
			var job entities.Job
			var ok bool

			select {
			case <-ctx.Done():
				return
			case job, ok = <-service.Jobs:
				if !ok {
					return
				}
			}

			w := workerPool.Get(job.Host)
			select {
			case <-ctx.Done():
				return
			case w.jobs <- job:
			}
		}
	}()

	wg.Wait()
}
