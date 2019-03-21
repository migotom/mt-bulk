package schema

import (
	"fmt"
	"net"
	"strconv"
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

func (h Host) String() string {
	return fmt.Sprintf("ID:%d IP:%s Port:%s User:%s Pass:%s", h.ID, h.IP, h.Port, h.User, h.Pass)
}

// Hosts defines list of hosts to use.
type Hosts struct {
	hosts []Host
}

func (h *Hosts) parseHost(oldHost Host) (Host, error) {
	newHost := oldHost

	list := strings.Split(oldHost.IP, ":")
	if len(list) > 2 {
		// does not meet format IP:PORT
		return Host{}, fmt.Errorf(fmt.Sprintf("host invalid format: %s", oldHost.IP))
	}

	if len(list) == 2 {
		port, err := strconv.Atoi(list[1])
		if err != nil || port < 0 || port > 65535 {
			return Host{}, fmt.Errorf("port invalid format: %s", list[1])
		}
		newHost.Port = list[1]
	}

	// TODO refactor for range hosts, e.g. 192.168.1.1-192.168.1.100
	if ipaddr, err := net.ResolveIPAddr("ip", list[0]); err == nil {
		newHost.IP = ipaddr.IP.String()
		return newHost, nil
	}

	if IP, _, err := net.ParseCIDR(oldHost.IP); err == nil {
		newHost.IP = IP.String()
		return newHost, nil
	}

	return Host{}, fmt.Errorf(fmt.Sprintf("can't resolve host: %s", oldHost.IP))
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
		return fmt.Errorf("hosts loader add %v", err)
	}

	h.hosts = append(h.hosts, hosts...)
	return nil
}
