package service

import (
	"context"
	"sync"

	"github.com/migotom/mt-bulk/internal/entities"
)

// Service of set of workers processing jobs on provided Mikrotik devices.
type Service struct {
	config Config
	Jobs   chan entities.Job
}

// NewService returns new service.
func NewService(config Config) *Service {
	return &Service{
		config: config,
		Jobs:   make(chan entities.Job, config.Workers),
	}
}

// Listen runs worker pool of SSH/Mikrotik SSL API clients processing jobs channel and sending jobs execution result into separate channel.
func (service *Service) Listen(ctx context.Context) {
	wg := new(sync.WaitGroup)

	workerPool := NewWorkerPool(service.config.Workers)
	for i := 0; i < service.config.Workers; i++ {
		w := NewWorker(8, service.config.Version)
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
		for {
			select {
			case <-ctx.Done():
				return
			case job, ok := <-service.Jobs:
				if !ok {
					return
				}
				w := workerPool.Get(job.Host)
				w.jobs <- job
			}
		}
	}()

	wg.Wait()
}
