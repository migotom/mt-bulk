package mtbulkrestapi

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/kvdb"
	"github.com/migotom/mt-bulk/internal/service"
)

// MTbulkRESTGateway service.
type MTbulkRESTGateway struct {
	Service *service.Service
	kv      kvdb.KV
	sugar   *zap.SugaredLogger

	Config
}

// NewMTbulkRestGateway returns new MTbulkRESTGateway service.
func NewMTbulkRestGateway(sugar *zap.SugaredLogger, arguments map[string]interface{}, version string) (*MTbulkRESTGateway, error) {
	config, err := configParser(arguments, version)
	if err != nil {
		return &MTbulkRESTGateway{}, err
	}

	kv, err := kvdb.OpenKV(sugar, config.Service.KVStore)
	if err != nil {
		return &MTbulkRESTGateway{}, err
	}

	return &MTbulkRESTGateway{
		sugar:   sugar,
		kv:      kv,
		Config:  config,
		Service: service.NewService(sugar, kv, config.Service),
	}, nil
}

// RunWorkers runs service workers and process all provided jobs.
// Returns after process of all jobs.
func (mtbulk *MTbulkRESTGateway) RunWorkers(ctx context.Context, cancel context.CancelFunc) {
	defer mtbulk.kv.Close()

	mtbulk.Service.Listen(ctx, cancel)
}

// JobHandler parses job request and process it with pool of workers.
func (mtbulk *MTbulkRESTGateway) JobHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var job entities.Job

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&job); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		job.Host.Parse()

		if err := mtbulk.AuthorizeRequest(r, &job); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		id := r.Context().Value("id").(string)
		mtbulk.sugar.Infow("processing job", "commands", job.Commands, "id", id)

		resultChan := make(chan entities.Result)
		if job.Data == nil {
			job.Data = make(map[string]string)
		}
		job.Data["root_directory"] = mtbulk.RootDirectory
		job.Result = resultChan
		job.ID = id

		// send job to workers
		select {
		case <-r.Context().Done():
			http.Error(w, "request cancelled by host", http.StatusGone)
			return
		case mtbulk.Service.Jobs <- job:
		}

		// fetch result
		select {
		case <-r.Context().Done():
			http.Error(w, "request cancelled by host", http.StatusGone)
			return
		case result := <-resultChan:
			if result.Errors != nil {
				w.WriteHeader(http.StatusNotAcceptable)
			}

			if err := json.NewEncoder(w).Encode(&result); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}
