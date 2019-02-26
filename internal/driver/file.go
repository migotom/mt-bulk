package driver

import (
	"bufio"
	"os"

	"github.com/migotom/mt-bulk/internal/schema"
)

// FileLoadHosts loads list of hosts from file
func FileLoadHosts(hostParser schema.HostParserFunc, filename string) ([]schema.Host, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []schema.Host
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		host, err := hostParser(schema.Host{IP: scanner.Text()})
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}
