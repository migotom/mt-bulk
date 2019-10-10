package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/migotom/mt-bulk/internal/entities"
)

// Service of set of workers processing jobs on provided Mikrotik devices.
type Service struct {
	config  Config
	Results chan entities.Result
	Jobs    chan entities.Job
}

// NewService returns new service.
func NewService(config Config) *Service {
	return &Service{
		config:  config,
		Results: make(chan entities.Result, config.Workers),
		Jobs:    make(chan entities.Job, config.Workers),
	}
}

// Listen runs worker pool of SSH/Mikrotik SSL API clients processing jobs channel and sending jobs execution result into separate channel.
func (service *Service) Listen(ctx context.Context) {
	wg := new(sync.WaitGroup)

	if !service.config.SkipVersionCheck {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := checkVersion(service.config.Version); err != nil {
				service.Results <- entities.Result{
					Host:  entities.Host{IP: "github.com", User: "public", Port: "443"},
					Job:   entities.Job{Kind: "CheckMTbulkVersion"},
					Error: fmt.Errorf("[Warrning] %s", err),
				}
			}
		}()
	}

	workerPool := NewWorkerPool(service.config.Workers)
	for i := 0; i < service.config.Workers; i++ {
		w := NewWorker(8, service.Results)
		workerPool.Add(w)

		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()

			w.ProcessJobs(ctx, service.config.Clients)
		}(w)
	}

	wg.Add(1)
	go func() {
		defer workerPool.Close()
		defer wg.Done()

		// jobs dispatcher
		for job := range service.Jobs {
			w := workerPool.Get(job.Host)
			w.jobs <- job
		}
	}()

	wg.Wait()
	close(service.Results)
}
