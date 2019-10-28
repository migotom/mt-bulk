package vulnerabilities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CVE defines single Common Vulnerabilities and Exposures.
type CVE struct {
	ID       string  `json:"id"`
	CVSS     float32 `json:"cvss"`
	Modified string  `json:"modified"`
	Summary  string  `json:"summary"`

	References []string `json:"references"`
}

func (cve CVE) String() string {
	var description strings.Builder
	description.WriteString(fmt.Sprintf("%s (CVSS %.1f): %s\n", cve.ID, cve.CVSS, cve.Summary))

	for _, ref := range cve.References {
		description.WriteString(fmt.Sprintf("- %s\n", ref))
	}
	return description.String()
}

// CVEsDownload downloads actual list of CVEs from exterpan public cve-search repository.
func (vm *Manager) CVEsDownload(ctx context.Context) error {
	vm.sugar.Infof("Downloading CVEs")

	if vm.cvesURL == "" {
		return errors.New("Missing CVEs download URL")
	}

	var res *http.Response
	var err error

	client := &http.Client{}
	for range []int{1, 2, 3} {
		var request *http.Request
		request, err = http.NewRequest("GET", vm.cvesURL, nil)
		if err != nil {
			return err
		}
		request.Header.Set("Accept", "application/json")
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/70.0")
		request = request.WithContext(ctx)
		res, err = client.Do(request)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response status code %v", res.StatusCode)
	}

	// find begin of stream with Data objects
	dec := json.NewDecoder(res.Body)
	for {
		t, err := dec.Token()
		if err != nil {
			return fmt.Errorf("invalid response body %v", err)
		}

		delim, ok := t.(json.Delim)
		if !ok {
			continue
		}
		if delim == '[' {
			break
		}
	}

	txn := vm.kv.NewTransaction()
	defer txn.Discard()

	// extract CVEs from JSON and store in K/V db
	versionsCVEs := make(map[int][]string)
	for dec.More() {
		var cve struct {
			ID       string  `json:"id"`
			CVSS     float32 `json:"cvss"`
			Modified string  `json:"Modified"`
			Summary  string  `json:"summary"`

			References               []string `json:"references"`
			VulnerableConfigurations []string `json:"vulnerable_configuration"`
		}
		err := dec.Decode(&cve)
		if err != nil {
			return errors.New("couldn't decode cve record")
		}

		err = txn.Store(fmt.Sprintf("%s%s", kvTagCVE, cve.ID), CVE{
			ID:         cve.ID,
			CVSS:       cve.CVSS,
			Modified:   cve.Modified,
			Summary:    cve.Summary,
			References: cve.References,
		})
		if err != nil {
			return err
		}

		for _, version := range ConfigurationsToVersions(cve.VulnerableConfigurations) {
			versionsCVEs[version] = append(versionsCVEs[version], cve.ID)
		}
	}

	for version, cves := range versionsCVEs {
		err = txn.Store(fmt.Sprintf("%s%d", kvTagVersion, version), cves)
		if err != nil {
			return err
		}
	}

	err = txn.Store(kvTagDBLastUpdate, time.Now())
	if err != nil {
		return err
	}

	err = txn.Store(kvTagDBVersion, RequiredKVDBVersion)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}
