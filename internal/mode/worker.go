package mode

import (
	"fmt"
	"log"
	"sync"

	"github.com/migotom/mt-bulk/internal/schema"
)

// Worker grabs devices to handle with configured handler from hosts channel.
func Worker(appConfig *schema.GeneralConfig, hosts chan schema.Host, errors chan schema.Error, wg *sync.WaitGroup) {
	defer wg.Done()
	for host := range hosts {
		if err := appConfig.ModeHandler(appConfig, host); err != nil {
			if appConfig.IgnoreErrors {
				log.Println(fmt.Errorf("[IP:%s][error] %s", host.IP, err))
				errors <- schema.Error{Host: host, Message: err.Error()}
			} else {
				log.Panicln(fmt.Errorf("[IP:%s][error] %s", host.IP, err))
			}
		}
	}
}

func ErrorCollector(appConfig *schema.GeneralConfig, errors chan schema.Error, wg *sync.WaitGroup) {
	defer wg.Done()

	hostsErrors := make(map[schema.Host][]string)
	for error := range errors {
		hostsErrors[error.Host] = append(hostsErrors[error.Host], error.Message)
	}

	if len(hostsErrors) == 0 {
		return
	}

	if appConfig.SkipSummary {
		return
	}

	fmt.Println()
	fmt.Println("Errors list:")
	for host, errors := range hostsErrors {
		fmt.Printf("Device: %s:%s\n", host.IP, host.Port)

		for _, error := range errors {
			fmt.Printf("\t%s\n", error)
		}
	}
}
