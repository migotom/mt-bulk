package main

import (
	"log"
	"os"

	docopt "github.com/docopt/docopt-go"
	mtbulk "github.com/migotom/mt-bulk/internal/mt-bulk"
	"github.com/migotom/mt-bulk/internal/schema"
)

var usage = `MT-bulk.

Usage:
  mt-bulk gen-certs [options]
  mt-bulk init-secure-api [options] [<hosts>...]
  mt-bulk change-password (--new=<newpass>) [options] [<hosts>...]  
  mt-bulk custom-api [--commands-file=<commands>] [options] [<hosts>...]  
  mt-bulk custom-ssh [--commands-file=<commands>] [options] [<hosts>...]  
  mt-bulk -h | --help
  mt-bulk --version

Options:
  -C <config-file>         Use configuration file, e.g. certs locations, ports, commands sequences, custom commands, etc...
  -s                       Be quiet and don't print commands and commands results to standard output
  -w <workers>             Number of parallel connections to run (default: 4)
  --skip-version-check     Skip checking for new version
  --skip-summary           Skip errors summary
  --exit-on-error          In case of any error stop executing commands
  --source-db              Load hosts using database configured by -C <config-file>
  --source-file=<file-in>  Load hosts from file <file-in>
`

const version = "1.5"

func main() {
	arguments, _ := docopt.ParseArgs(usage, os.Args[1:], version)
	//	fmt.Println(arguments)

	appConfig := schema.GeneralConfig{Version: version}

	hostsLoaders, _, err := configParser(arguments, &appConfig)
	if err != nil {
		log.Fatalf("[Fatal error] Config parser: %s\n", err)
	}

	service := mtbulk.NewService(&appConfig, hostsLoaders)
	os.Exit(service.Run())
}
