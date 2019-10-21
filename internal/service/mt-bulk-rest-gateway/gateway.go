package mtbulkrestapi

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/service"
)

// MTbulkRESTGateway service.
type MTbulkRESTGateway struct {
	Service *service.Service
	sugar   *zap.SugaredLogger

	Config
}

// NewMTbulkRestGateway returns new MTbulkRESTGateway service.
func NewMTbulkRestGateway(sugar *zap.SugaredLogger, arguments map[string]interface{}, version string) (*MTbulkRESTGateway, error) {
	config, err := configParser(arguments, version)
	if err != nil {
		return &MTbulkRESTGateway{}, err
	}

	return &MTbulkRESTGateway{
		sugar:   sugar,
		Config:  config,
		Service: service.NewService(sugar, config.Service),
	}, nil
}

// RunWorkers runs service workers and process all provided jobs.
// Returns after process of all jobs.
func (mtbulk *MTbulkRESTGateway) RunWorkers(ctx context.Context) {
	mtbulk.Service.Listen(ctx)
}

// JobHandler parses job request and process it with pool of workers.
func (mtbulk *MTbulkRESTGateway) JobHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var job entities.Job

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&job); err != nil {
			http.Error(w, "Bad request", 400)
			return
		}
		job.Host.Parse()

		if err := mtbulk.AuthorizeRequest(r, &job); err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		id := r.Context().Value("id").(string)
		mtbulk.sugar.Infow("processing job", "commands", job.Commands, "id", id)

		resultChan := make(chan entities.Result)
		job.Data["root_directory"] = mtbulk.RootDirectory
		job.Result = resultChan
		job.ID = id

		// send job to workers
		select {
		case <-r.Context().Done():
			http.Error(w, "Request cancelled by host", 410)
			return
		case mtbulk.Service.Jobs <- job:
		}

		// fetch result
		select {
		case <-r.Context().Done():
			http.Error(w, "Request cancelled by host", 410)
			return
		case result := <-resultChan:
			if result.Error != nil {
				w.WriteHeader(http.StatusNotAcceptable)
			}

			if err := json.NewEncoder(w).Encode(&result); err != nil {
				http.Error(w, err.Error(), 500)
			}
		}
	}
}
