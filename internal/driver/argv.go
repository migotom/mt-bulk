package driver

import (
	"github.com/migotom/mt-bulk/internal/schema"
)

// ArgvLoadHosts loads lists of hosts using standard argument list.
func ArgvLoadHosts(hostParser schema.HostParserFunc, data []string) (hosts []schema.Host, err error) {
	hosts = make([]schema.Host, len(data))
	for i, entry := range data {
		if hosts[i], err = hostParser(schema.Host{IP: entry}); err != nil {
			return nil, err
		}
	}
	return hosts, nil
}
