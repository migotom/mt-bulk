package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	docopt "github.com/docopt/docopt-go"
	mtbulk "github.com/migotom/mt-bulk/internal/service/mt-bulk"
	"go.uber.org/zap"
)

var usage = `MT-bulk.

Usage:
  mt-bulk gen-api-certs [options]
  mt-bulk gen-ssh-keys [options]
  mt-bulk init-secure-api [options] [<hosts>...]
  mt-bulk init-publickey-ssh [options] [<hosts>...]
  mt-bulk change-password (--new=<newpass>) [--user=<user>] [options] [<hosts>...]  
  mt-bulk system-backup (--name=<name>) (--backup-store=<backups>) [options] [<hosts>...]  
  mt-bulk sftp <source> <target> [options] [<hosts>...]  
  mt-bulk custom-api [--commands-file=<commands>] [options] [<hosts>...]  
  mt-bulk custom-ssh [--commands-file=<commands>] [options] [<hosts>...]  
  mt-bulk security-audit [options] [<hosts>...] 
  mt-bulk -h | --help
  mt-bulk --version

Options:
  -C <config-file>         Use configuration file, e.g. certs locations, ports, commands sequences, custom commands, etc...
  --source-db              Load hosts using database configured by -C <config-file>
  --source-file=<file-in>  Load hosts from file <file-in>
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

	mtbulk, err := mtbulk.NewMTbulk(sugar, arguments, version)
	if err != nil {
		log.Fatalf("Configuration parser error: %s\n", err)
	}

	wg := new(sync.WaitGroup)

	// gracefull exit
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		sig := <-signals
		log.Printf("Interrupted by signal: %v\n", sig)

		cancel()
	}()

	// load jobs and hosts
	wg.Add(1)
	go func() {
		defer wg.Done()

		mtbulk.LoadJobs(ctx)
	}()

	// collect and present responses
	wg.Add(1)
	go func() {
		defer wg.Done()

		mtbulk.ResponseCollector(ctx)
	}()

	// run workers
	mtbulk.Listen(ctx, cancel)

	wg.Wait()
	os.Exit(mtbulk.Status.Get())
}
