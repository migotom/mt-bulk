package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	docopt "github.com/docopt/docopt-go"
	"github.com/migotom/mt-bulk/internal/mode"
	"github.com/migotom/mt-bulk/internal/schema"
)

var usage = `MT-bulk.

Usage:
  mt-bulk custom-api [--custom-args=<custom-args>] [options] [<hosts>...]  
  mt-bulk custom-ssh [--custom-args=<custom-args>] [options] [<hosts>...]  
  mt-bulk init-secure-api [--gen-certs] [options] [<hosts>...]
  mt-bulk change-password (--new=<newpass>) [options] [<hosts>...]  
  mt-bulk -h | --help
  mt-bulk --version

Options:
  -C <config-file>         Use configuration file, eg. certs locations, ports, commands sequences, custom commands, etc...
  -s                       Be quiet and don't print commands and commands results to standard output
  -w <workers>             Number of paralell connections to run (default: 4)
  --skip-summary           Skip errors summary
  --exit-on-error          In case of any error stop executing commands
  --source-db              Load hosts using database configured by -C <config-file>
  --source-file=<file-in>  Load hosts from file <file-in>
`

const version = "0.1.0"

func loadHosts(hostsLoaders *[]schema.HostsLoaderFunc, hosts *schema.Hosts) {
	hosts.Reset()
	for _, hostsLoader := range *hostsLoaders {
		if err := hosts.Add(hostsLoader); err != nil {
			log.Fatal(fmt.Errorf("loadHosts fatal error: %v", err))
		}
	}
}

func main() {
	arguments, _ := docopt.ParseArgs(usage, os.Args[1:], version)
	//fmt.Println(arguments)

	appConfig := schema.GeneralConfig{}

	hostsLoaders, _, err := configParser(arguments, &appConfig)
	if err != nil {
		log.Fatalf("[Fatal error] %s\n", err)
	}

	var Hosts schema.Hosts
	loadHosts(&hostsLoaders, &Hosts)
	if len(Hosts.Get()) == 0 {
		log.Fatalln("No hosts to process.")
	}

	hostsChan := make(chan schema.Host, appConfig.Workers)
	resultsChan := make(chan schema.Error, appConfig.Workers)

	var wgWorkers sync.WaitGroup
	var wgCollector sync.WaitGroup

	// TODO add graceful shutdown

	wgCollector.Add(1)
	go mode.ErrorCollector(&appConfig, resultsChan, &wgCollector)

	for i := 1; i <= appConfig.Workers; i++ {
		wgWorkers.Add(1)
		go mode.Worker(&appConfig, hostsChan, resultsChan, &wgWorkers)
	}

	for _, host := range Hosts.Get() {
		hostsChan <- host
	}

	close(hostsChan)
	wgWorkers.Wait()

	close(resultsChan)
	wgCollector.Wait()
}
