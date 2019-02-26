package schema

import (
	"fmt"
	"net"
	"strings"
)

// HostParserFunc ...
type HostParserFunc func(Host) (Host, error)

// HostsLoaderFunc ...
type HostsLoaderFunc func(HostParserFunc) ([]Host, error)

// HostsCleanerFunc cleanups handlers, connections, open sockets, files etc. used by Loader/Saver/Parser.
type HostsCleanerFunc func()

type Host struct {
	ID   int
	IP   string
	Port string
	User string
	Pass string
}

// Hosts defines list of hosts to use.
type Hosts struct {
	hosts []Host
}

func (h *Hosts) parseHost(oldHost Host) (Host, error) {
	newHost := oldHost

	list := strings.Split(oldHost.IP, ":")
	if len(list) > 2 {
		return Host{}, fmt.Errorf(fmt.Sprintf("Host invalid format: %s", oldHost.IP))
	}

	if len(list) == 2 {
		newHost.Port = list[1]
	}

	//TODO
	//refactor for range hosts
	ipaddr, err := net.ResolveIPAddr("ip", list[0])
	if err == nil {
		newHost.IP = ipaddr.IP.String()
		return newHost, nil
	}

	IP, _, err := net.ParseCIDR(oldHost.IP)
	if err == nil {
		newHost.IP = IP.String()
		return newHost, nil
	}

	return Host{}, fmt.Errorf(fmt.Sprintf("Can't resolve host: %s", oldHost.IP))
}

// Get list of hosts.
func (h *Hosts) Get() []Host {
	return h.hosts
}

// Reset list of hosts.
func (h *Hosts) Reset() {
	h.hosts = nil
}

// Add hosts using HostsLoader function.
func (h *Hosts) Add(loader HostsLoaderFunc) error {
	hosts, err := loader(h.parseHost)
	if err != nil {
		return fmt.Errorf("Hosts loader Add %v", err)
	}

	for _, host := range hosts {
		h.hosts = append(h.hosts, host)
	}
	return nil
}
