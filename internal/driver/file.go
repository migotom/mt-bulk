package driver

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

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
