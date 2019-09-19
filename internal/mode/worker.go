package mode

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// Worker grabs devices to handle with configured handler from hosts channel.
func Worker(ctx context.Context, appConfig *schema.GeneralConfig, hosts chan schema.Host, errors chan schema.Error) {
	for {
		select {
		case <-ctx.Done():
			return
		case host, ok := <-hosts:
			if !ok {
				return
			}
			if err := appConfig.ModeHandler(ctx, appConfig, host); err != nil {
				if appConfig.IgnoreErrors {
					log.Println(fmt.Errorf("[IP:%s][error] %s", host.IP, err))
					errors <- schema.Error{Host: host, Message: err.Error()}
				} else {
					log.Panicln(fmt.Errorf("[IP:%s][error] %s", host.IP, err))
				}
			}
		}
	}
}

// ErrorCollector collects and parde all errors produced by workers.
func ErrorCollector(appConfig *schema.GeneralConfig, errors chan schema.Error, status *service.ApplicationStatus, wg *sync.WaitGroup) {
	defer wg.Done()

	hostsErrors := make(map[schema.Host][]string)
	for error := range errors {
		hostsErrors[error.Host] = append(hostsErrors[error.Host], error.Message)
	}

	if len(hostsErrors) == 0 {
		return
	}

	status.SetCode(1)

	if appConfig.SkipSummary {
		return
	}

	fmt.Println()
	fmt.Println("Errors list:")
	for host, errors := range hostsErrors {
		if host.IP != "" {
			fmt.Printf("Device: %s:%s\n", host.IP, host.Port)
		}

		for _, error := range errors {
			fmt.Printf("\t%s\n", error)
		}
	}
}
