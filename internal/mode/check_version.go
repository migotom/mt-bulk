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

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"go.uber.org/zap"
)

// CheckMTbulkVersion executes by client custom job.
func CheckMTbulkVersion(version string) OperationModeFunc {
	return func(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) ([]entities.CommandResult, error) {
		if err := checkVersion(version); err != nil {
			return []entities.CommandResult{
				entities.CommandResult{
					Body:  "/<mt-bulk>check version",
					Error: err,
				},
			}, err
		}
		return nil, nil
	}
}

func checkVersion(currentVersion string) error {
	type release struct {
		Draft   bool
		URL     string `json:"html_url"`
		Version string `json:"name"`
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.github.com/repos/migotom/mt-bulk/releases/latest", nil)
	if err != nil {
		return fmt.Errorf("can't create request to fetch latest release info: %s", err)
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("can't fetch latest release info: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("can't fetch latest release info, status code: %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("can't read details of latest release: %s", err)
	}

	var currentRelease release
	if err := json.Unmarshal(body, &currentRelease); err != nil {
		return fmt.Errorf("can't parse details of latest release: %s", err)
	}

	if currentRelease.Draft {
		return nil
	}

	currentVersionInt, _ := parseVersion(currentVersion)
	releasedVersionInt, err := parseVersion(currentRelease.Version)

	if err != nil {
		return fmt.Errorf("invalid version number in latest release: %s", err)
	}

	if currentVersionInt < releasedVersionInt {
		return fmt.Errorf("new version of MT-bulk v%v available at %v", currentRelease.Version, currentRelease.URL)
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
