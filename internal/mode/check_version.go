package mode

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/kvdb"
)

// CheckMTbulkVersion executes by client custom job.
func CheckMTbulkVersion(version string, kv kvdb.KV) OperationModeFunc {
	return func(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) entities.Result {
		if err := checkVersion(version, kv); err != nil {
			return entities.Result{
				Results: []entities.CommandResult{
					entities.CommandResult{
						Body:  "/<mt-bulk>check version",
						Error: err,
					},
				},
				Errors: []error{err}}
		}
		return entities.Result{}
	}
}

func checkVersion(currentVersion string, kv kvdb.KV) error {
	type release struct {
		Draft   bool
		URL     string `json:"html_url"`
		Version string `json:"name"`
	}

	var releasedVersion release
	var lastUpdate time.Time

	_ = kv.View(func(txn kvdb.Txn) error {
		_ = txn.GetCopy("MTbulk::LastUpdate", &lastUpdate)
		_ = txn.GetCopy("MTbulk::ReleasedVersion", &releasedVersion)
		return nil
	})

	if lastUpdate.Before(time.Now().Add(-24 * time.Duration(time.Hour))) {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		req, err := http.NewRequest("GET", "https://api.github.com/repos/migotom/mt-bulk/releases/latest", nil)
		if err != nil {
			return nil
		}

		req.Header.Add("Accept", "application/vnd.github.v3+json")
		resp, err := client.Do(req)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			return nil
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil
		}

		if err := json.Unmarshal(body, &releasedVersion); err != nil {
			return nil
		}

		if releasedVersion.Draft {
			return nil
		}

		txn := kv.NewTransaction()
		defer txn.Discard()

		err = txn.Store("MTbulk::LastUpdate", time.Now())
		if err != nil {
			return err
		}

		err = txn.Store("MTbulk::ReleasedVersion", &releasedVersion)
		if err != nil {
			return err
		}
		txn.Commit()
	}

	releasedVersionInt, err := parseVersion(releasedVersion.Version)
	if err != nil {
		return fmt.Errorf("invalid version number in latest release: %s", err)
	}
	currentVersionInt, _ := parseVersion(currentVersion)
	if currentVersionInt < releasedVersionInt {
		return fmt.Errorf("new version of MT-bulk v%v available at %v", releasedVersion.Version, releasedVersion.URL)
	}

	return nil
}

func parseVersion(version string) (result int64, err error) {
	if matches := regexp.MustCompile(`(\d+)\.(\d+)(?:\.(\d+))?`).FindStringSubmatch(version); len(matches) > 1 {
		v := ""
		for i := 1; i <= 3; i++ {
			v = fmt.Sprintf("%s%04s", v, matches[i])
		}

		if result, err = strconv.ParseInt(v, 10, 64); err != nil {
			return 0, err
		}
	}
	return
}
