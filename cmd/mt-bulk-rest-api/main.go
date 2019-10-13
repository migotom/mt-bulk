package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	docopt "github.com/docopt/docopt-go"
	"github.com/gorilla/mux"
	mtbulkrestapi "github.com/migotom/mt-bulk/internal/service/mt-bulk-rest-gateway"
)

var usage = `MT-bulk REST gateway.

Usage:
  mt-bulk-rest-gw [options]
  mt-bulk-rest-gw gen-https-certs [options]
	mt-bulk-rest-gw gen-refresh-token [options]
  mt-bulk -h | --help
  mt-bulk --version

Options:
  -C <config-file>         Use configuration file, e.g. certs locations, ports, commands sequences, custom commands, etc...
`

var version string

func main() {
	arguments, _ := docopt.ParseArgs(usage, os.Args[1:], version)
	ctx, cancel := context.WithCancel(context.Background())

	mtbulkRESTAPI, err := mtbulkrestapi.NewMTbulkRestGateway(arguments, version)
	if err != nil {
		log.Fatalf("Configuration parser error: %s\n", err)
	}

	router := mux.NewRouter()
	router.Use(commonMiddleware)
	router.HandleFunc("/authenticate", mtbulkRESTAPI.Authenticate(ctx)).Methods("POST")

	jobRouter := router.PathPrefix("/job").Subrouter()
	jobRouter.Use(mtbulkRESTAPI.AuthorizeMiddleware)
	jobRouter.HandleFunc("", mtbulkRESTAPI.JobHandler(ctx)).Methods("POST")

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	wg := new(sync.WaitGroup)

	// gracefull exit
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		sig := <-signals
		log.Printf("Interrupted by signal: %v\n", sig)

		cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Fatalf("HTTP server shutdown %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// run workers
	mtbulkRESTAPI.RunWorkers(ctx)
	wg.Wait()
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
