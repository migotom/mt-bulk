package vulnerabilities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/migotom/mt-bulk/internal/kvdb"
	"go.uber.org/zap"
)

// CVEURLs defines information about CVE search engine endpoints.
type CVEURLs struct {
	DBInfo string `toml:"db_info" yaml:"db_info"`
	DB     string `toml:"db" yaml:"db"`
}

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

// CVEsDownload downloads actual list of CVEs from external public cve-search repository.
func (vm *Manager) CVEsDownload(ctx context.Context) error {
	txn := vm.kv.NewTransaction()
	defer txn.Discard()

	var err error
	for _, cveURLs := range vm.cvesURLs {
		var dbUpToDate bool
		var dbInfo cveDBInfo
		dbUpToDate, dbInfo, err = isDBinfoUpToDate(ctx, vm.sugar, vm.kv, cveURLs.DBInfo)
		if err != nil {
			continue
		}

		if dbUpToDate {
			return nil
		}

		err = downloadCVEs(ctx, vm.sugar, txn, cveURLs.DB)
		if err != nil {
			continue
		}

		err = txn.Store(kvTagDBCVEdbInfo, dbInfo)
		if err != nil {
			return err
		}

		txn.Commit()
		return nil
	}
	return err
}

func downloadCVEs(ctx context.Context, sugar *zap.SugaredLogger, txn kvdb.Txn, url string) error {
	res, err := downloadWithRetries(ctx, sugar, url)
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

	return nil
}

func isDBinfoUpToDate(ctx context.Context, sugar *zap.SugaredLogger, kv kvdb.KV, url string) (bool, cveDBInfo, error) {
	res, err := downloadWithRetries(ctx, sugar, url)
	if err != nil {
		return false, cveDBInfo{}, err
	}

	if res.StatusCode != http.StatusOK {
		return false, cveDBInfo{}, fmt.Errorf("invalid response status code %v", res.StatusCode)
	}

	defer res.Body.Close()

	dbInfoRemote := cveDBInfo{}
	err = json.NewDecoder(res.Body).Decode(&dbInfoRemote)
	if err != nil {
		return false, cveDBInfo{}, err
	}

	dbInfoLocal := cveDBInfo{}
	_ = kv.View(func(txn kvdb.Txn) error {
		return txn.GetCopy(kvTagDBCVEdbInfo, &dbInfoLocal)
	})

	return !dbInfoLocal.Before(dbInfoRemote), dbInfoRemote, nil
}

type cveDBInfo struct {
	CAPEC    cveDBInfoEntry `json:"capec"`
	CPE      cveDBInfoEntry `json:"cpe"`
	CPEOther cveDBInfoEntry `json:"cpeOther"`
	CVES     cveDBInfoEntry `json:"cves"`
	CWE      cveDBInfoEntry `json:"cwe"`
	VIA4CVE  cveDBInfoEntry `json:"via4"`
}

func (l cveDBInfo) Before(r cveDBInfo) bool {
	if l.CAPEC.LastUpdate.Time.Before(r.CAPEC.LastUpdate.Time) ||
		l.CPE.LastUpdate.Time.Before(r.CAPEC.LastUpdate.Time) ||
		l.CPEOther.LastUpdate.Time.Before(r.CPEOther.LastUpdate.Time) ||
		l.CVES.LastUpdate.Time.Before(r.CVES.LastUpdate.Time) ||
		l.CWE.LastUpdate.Time.Before(l.CWE.LastUpdate.Time) ||
		l.VIA4CVE.LastUpdate.Time.Before(l.VIA4CVE.LastUpdate.Time) {
		return true
	}
	return false
}

type cveDBInfoEntry struct {
	LastUpdate cveTime `json:"last_update"`
}
type cveTime struct {
	time.Time
}

func (t *cveTime) UnmarshalJSON(b []byte) error {
	t.Time, _ = time.Parse(time.RFC3339, fmt.Sprintf("%sZ", strings.ReplaceAll(string(b), "\"", "")))
	return nil
}
