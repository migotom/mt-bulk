package driver

import (
	"github.com/migotom/mt-bulk/internal/schema"
)

// ArgvLoadHosts loads lists of hosts using standard argument list.
func ArgvLoadHosts(hostParser schema.HostParserFunc, data []string) ([]schema.Host, error) {
	hosts := make([]schema.Host, len(data))
	for i, entry := range data {
		var err error

		hosts[i], err = hostParser(schema.Host{IP: entry})
		if err != nil {
			return nil, err
		}
	}

	return hosts, nil
}
