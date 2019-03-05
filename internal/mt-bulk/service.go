package mtbulk

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/migotom/mt-bulk/internal/mode"
	"github.com/migotom/mt-bulk/internal/schema"
)

type Service struct {
	appConfig        *schema.GeneralConfig
	hostsLoaderFuncs []schema.HostsLoaderFunc
	hosts            schema.Hosts

	hostsChannel   chan schema.Host
	resultsChannel chan schema.Error

	wgWorkers   sync.WaitGroup
	wgCollector sync.WaitGroup
	cancel      context.CancelFunc
}

func NewService(appConfig *schema.GeneralConfig, hostsLoaderFuncs []schema.HostsLoaderFunc) *Service {
	return &Service{
		appConfig:        appConfig,
		hostsLoaderFuncs: hostsLoaderFuncs,

		hostsChannel:   make(chan schema.Host, appConfig.Workers),
		resultsChannel: make(chan schema.Error, appConfig.Workers),
	}
}

func (s *Service) Start() error {
	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case sig := <-signals:
			log.Printf("Interrupted by signal: %v\n", sig)
		}
		s.cancel()
	}()

	if err := s.loadHosts(); err != nil {
		log.Fatalf("[Fatal error] loading hosts error %s\n", err)
	}

	s.wgCollector.Add(1)
	go mode.ErrorCollector(s.appConfig, s.resultsChannel, &s.wgCollector)

	for i := 1; i <= s.appConfig.Workers; i++ {
		s.wgWorkers.Add(1)
		go mode.Worker(ctx, s.appConfig, s.hostsChannel, s.resultsChannel, &s.wgWorkers)
	}

	for _, host := range s.hosts.Get() {
		s.hostsChannel <- host
	}
	return nil
}

func (s *Service) Close() {
	defer s.cancel()

	close(s.hostsChannel)
	s.wgWorkers.Wait()

	close(s.resultsChannel)
	s.wgCollector.Wait()
}

func (s *Service) loadHosts() error {
	s.hosts.Reset()

	for _, hostsLoader := range s.hostsLoaderFuncs {
		if err := s.hosts.Add(hostsLoader); err != nil {
			return err
		}
	}
	if len(s.hosts.Get()) == 0 && len(s.hostsLoaderFuncs) > 0 {
		return errors.New("empty list of hosts")
	}

	return nil
}
