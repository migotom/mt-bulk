package driver

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gocarina/gocsv"
	"gopkg.in/yaml.v2"

	"github.com/BurntSushi/toml"
	"github.com/migotom/mt-bulk/internal/entities"
)

// FileLoadJobs loads list of jobs from file.
func FileLoadJobs(ctx context.Context, jobTemplate entities.Job, filename string) (jobs []entities.Job, err error) {
	var hosts struct {
		Host []entities.Host
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".toml":
		err = toml.Unmarshal(content, &hosts)
		if err != nil {
			return nil, err
		}
	case ".yml", ".yaml":
		err = yaml.Unmarshal(content, &hosts)
		if err != nil {
			return nil, err
		}
    case ".csv":
        type H struct {
            IP string `csv:"Addresses"`
            Type string `csv:"Type"`
        }
        var hs []H
        err = gocsv.UnmarshalBytes(content, &hs)
        if err != nil {
            return nil, err
        }
        for _, h := range hs {
            // csv export from The Dude contains all Devices, so filter out everything that is not RouterOS
            if h.Type == "RouterOS" {
                hosts.Host = append(hosts.Host, entities.Host{IP: h.IP})
            }
        }
	default:
		reader := bytes.NewReader(content)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			hosts.Host = append(hosts.Host, entities.Host{IP: scanner.Text()})
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	for _, host := range hosts.Host {
		job := jobTemplate
		job.Host = host
		if err := job.Host.Parse(); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
