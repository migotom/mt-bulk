package entities

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Host represents single host instance with all data and credentials required to connect to.
type Host struct {
	ID       string
	IP       string `toml:"ip" yaml:"ip"`
	Port     string `toml:"port" yaml:"port"`
	User     string `toml:"user" yaml:"user"`
	Password string `toml:"password" yaml:"password"`
}

// GetPasswords returns list of available passwords.
func (h Host) GetPasswords() (passwords []string) {
	for _, password := range strings.Split(h.Password, ",") {
		password = strings.TrimSpace(password)
		passwords = append(passwords, password)
	}
	return
}

// SetDefaults sets host's default values.
func (h *Host) SetDefaults(defaultPort, defaultUser, defaultPasswords string) {
	if h.Port == "" {
		h.Port = defaultPort
	}
	if h.User == "" {
		h.User = defaultUser
	}
	if h.Password == "" {
		h.Password = defaultPasswords
	}
}

// Parse host IP address and split into IP, Port attribute if required.
func (h *Host) Parse() error {
	list := strings.Split(h.IP, ":")

	if len(list) > 2 {
		return fmt.Errorf(fmt.Sprintf("host invalid format: %s, allowed single host IP:PORT or range IP:PORT-IP:PORT, where PORT is optional", h.IP))
	}

	if len(list) == 2 {
		port, err := strconv.Atoi(list[1])
		if err != nil || port < 0 || port > 65535 {
			return fmt.Errorf("port invalid format: %s", list[1])
		}
		h.Port = list[1]
	}

	if ipaddr, err := net.ResolveIPAddr("ip", list[0]); err == nil {
		h.IP = ipaddr.IP.String()
		return nil
	}

	if IP, _, err := net.ParseCIDR(h.IP); err == nil {
		h.IP = IP.String()
		return nil
	}

	return fmt.Errorf(fmt.Sprintf("can't resolve host: %s", h.IP))
}

func (h Host) String() string {
	return fmt.Sprintf("%s@%s:%s", h.User, h.IP, h.Port)
}
