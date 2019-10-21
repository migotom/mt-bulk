package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"

	docopt "github.com/docopt/docopt-go"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	mtbulkrestapi "github.com/migotom/mt-bulk/internal/service/mt-bulk-rest-gateway"
)

var usage = `MT-bulk REST API gateway.

Usage:
  mt-bulk-rest-gw [options]
  mt-bulk-rest-gw gen-https-certs [options]
  mt-bulk-rest-gw -h | --help
  mt-bulk-rest-gw --version

Options:
  -C <config-file>         Use configuration file
`

var version string

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Could not initialize logger: %s\n", err)
	}
	sugar := logger.Sugar()
	defer sugar.Sync()

	arguments, _ := docopt.ParseArgs(usage, os.Args[1:], version)
	ctx, cancel := context.WithCancel(context.Background())

	mtbulkRESTAPI, err := mtbulkrestapi.NewMTbulkRestGateway(sugar, arguments, version)
	if err != nil {
		log.Fatalf("Configuration parser error: %s\n", err)
	}

	wg := new(sync.WaitGroup)

	// define routes
	router := mux.NewRouter()
	router.Use(mtbulkRESTAPI.LogMiddleware(ctx))
	router.HandleFunc("/authenticate", mtbulkRESTAPI.AuthenticateToken(ctx)).Methods("POST")

	jobRouter := router.PathPrefix("/job").Subrouter()
	jobRouter.Use(mtbulkRESTAPI.AuthorizeMiddleware)
	jobRouter.HandleFunc("", mtbulkRESTAPI.JobHandler(ctx)).Methods("POST")

	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	httpServer := &http.Server{
		Addr:         mtbulkRESTAPI.Config.Listen,
		Handler:      router,
		TLSConfig:    tlsConfig,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	// gracefull exit
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		sig := <-signals
		sugar.Infow("interrupted", "signal", sig)

		cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			sugar.Fatalw("HTTP server shutdown", "error", err)
		}
	}()

	// start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()

		sugar.Infow("MT-bulk REST API", "listen", mtbulkRESTAPI.Config.Listen)
		err := httpServer.ListenAndServeTLS(
			filepath.Join(mtbulkRESTAPI.Config.KeyStore, "rest-api.crt"),
			filepath.Join(mtbulkRESTAPI.Config.KeyStore, "rest-api.key"),
		)
		if err != http.ErrServerClosed {
			sugar.Fatalw("HTTP server listen", "error", err)
		}
	}()

	// run workers
	mtbulkRESTAPI.RunWorkers(ctx)
	wg.Wait()
}
