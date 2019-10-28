package vulnerabilities

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/dgraph-io/badger"
	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/kvdb"
)

// Manager is vulnerability check worker.
type Manager struct {
	cvesURL string
	kv      kvdb.KV
	sugar   *zap.SugaredLogger

	jobs chan vulnerabilityCheckJob
}

// NewManager returns new vulnerability manager.
func NewManager(sugar *zap.SugaredLogger, cvesURL string, kv kvdb.KV) *Manager {
	return &Manager{
		sugar:   sugar,
		cvesURL: cvesURL,
		kv:      kv,
		jobs:    make(chan vulnerabilityCheckJob, 1),
	}
}

// Check checks asynchronously given version for any known CVE.
func (vm *Manager) Check(version string) error {
	err := make(chan error)

	vm.jobs <- vulnerabilityCheckJob{version: version, err: err}
	return <-err
}

// Listen to any job with specified version to check and response channel.
func (vm *Manager) Listen(ctx context.Context) {
	for {
		select {
		case job := <-vm.jobs:
			job.err <- vm.check(ctx, job.version)
			close(job.err)
		case <-ctx.Done():
			return
		}
	}
}

type vulnerabilityCheckJob struct {
	version string
	err     chan error
}

func (vm *Manager) check(ctx context.Context, input string) error {
	var lastUpdate time.Time
	var dbVersion int

	_ = vm.kv.View(func(txn kvdb.Txn) error {
		_ = txn.GetCopy(kvTagDBLastUpdate, &lastUpdate)
		_ = txn.GetCopy(kvTagDBVersion, &dbVersion)
		return nil
	})

	if lastUpdate.Before(time.Now().Add(-24*time.Duration(time.Hour))) || dbVersion < RequiredKVDBVersion {
		if err := vm.CVEsDownload(ctx); err != nil {
			return fmt.Errorf("can't download vulnerabilities: %v", err)
		}
	}

	versionToCheck := versionToInt(input)
	vulnerabilityError := VulnerabilityError{}

	err := vm.kv.View(func(txn kvdb.Txn) error {
		extractVersionRe := regexp.MustCompile(fmt.Sprintf("^%s(.*)", kvTagVersion))

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.KeyCopy(nil))

			match := extractVersionRe.FindStringSubmatch(key)
			if match == nil || len(match) != 2 {
				continue
			}

			versionNumber, err := strconv.Atoi(match[1])
			if err != nil {
				return err
			}

			if versionNumber < versionToCheck {
				continue
			}

			var list []string
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			if err = gob.NewDecoder(bytes.NewReader(value)).Decode(&list); err != nil {
				return err
			}

		cvescan:
			for _, cveID := range list {
				for _, noticedCve := range vulnerabilityError.Vulnerabilities {
					if cveID == noticedCve.ID {
						continue cvescan
					}
				}

				var cve CVE
				err := txn.GetCopy(fmt.Sprintf("%s%s", kvTagCVE, cveID), &cve)
				if err != nil {
					return err
				}
				vulnerabilityError.Vulnerabilities = append(vulnerabilityError.Vulnerabilities, cve)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if vulnerabilityError.Error() == "" {
		return nil
	}
	return vulnerabilityError
}
