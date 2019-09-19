package mtbulk

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/migotom/mt-bulk/internal/mode"
	"github.com/migotom/mt-bulk/internal/schema"
	"github.com/migotom/mt-bulk/internal/service"
)

// MTBulk represents MT-bulk as service.
type MTBulk struct {
	appConfig        *schema.GeneralConfig
	hostsLoaderFuncs []schema.HostsLoaderFunc
	hosts            []schema.Host
	Status           service.ApplicationStatus
}

// NewMTBulk builds new MT-bulk service.
func NewMTBulk(appConfig *schema.GeneralConfig, hostsLoaderFuncs []schema.HostsLoaderFunc) *MTBulk {
	return &MTBulk{
		appConfig:        appConfig,
		hostsLoaderFuncs: hostsLoaderFuncs,
	}
}

// Run service, run workers and prepare self to gracefull exit if needed.
func (mtbulk *MTBulk) Run() int {
	var ctx context.Context
	ctx, cancel := context.WithCancel(context.Background())

	hostsChannel := make(chan schema.Host, mtbulk.appConfig.Workers)
	resultsChannel := make(chan schema.Error, mtbulk.appConfig.Workers)

	wg := new(sync.WaitGroup)
	wgCollector := new(sync.WaitGroup)

	// check app version
	if !mtbulk.appConfig.SkipVersionCheck {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := checkVersion(mtbulk.appConfig.Version); err != nil {
				resultsChannel <- schema.Error{
					Message: fmt.Sprintf("[Warrning] %s", err),
				}
			}
		}()
	}

	// gracefull exit
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		sig := <-signals
		log.Printf("Interrupted by signal: %v\n", sig)

		cancel()
	}()

	if err := mtbulk.loadHosts(); err != nil {
		log.Fatalf("[Fatal error] loading hosts error %s\n", err)
	}

	wgCollector.Add(1)
	go func() {
		defer wgCollector.Done()
		mode.ErrorCollector(mtbulk.appConfig, resultsChannel, &mtbulk.Status)
	}()

	for i := 1; i <= mtbulk.appConfig.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mode.Worker(ctx, mtbulk.appConfig, hostsChannel, resultsChannel)
		}()
	}

	for _, host := range mtbulk.hosts {
		hostsChannel <- host
	}

	close(hostsChannel)
	wg.Wait()

	close(resultsChannel)
	wgCollector.Wait()

	return mtbulk.Status.Get()
}

func (mtbulk *MTBulk) loadHosts() (err error) {
	mtbulk.hosts = nil

	for _, hostsLoader := range mtbulk.hostsLoaderFuncs {
		if mtbulk.hosts, err = hostsLoader(schema.HostParser); err != nil {
			return
		}
	}
	if len(mtbulk.hosts) == 0 && len(mtbulk.hostsLoaderFuncs) > 0 {
		return errors.New("empty list of hosts")
	}

	return nil
}
