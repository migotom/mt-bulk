package vulnerabilities

import (
	"context"
	"errors"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// ConfigurationsToVersions converts list of CVEdefined vulnerable configurations into list of Mikrotik versions as integers.
func ConfigurationsToVersions(vulnerableConfigurations []string) []int {
	versions := make([]int, 0, len(vulnerableConfigurations))

	extractVersionRe := regexp.MustCompile(`:mi[ck]rotik:routeros:([\d\.]+):?`)

configurations:
	for _, configuration := range vulnerableConfigurations {
		versionMatch := extractVersionRe.FindStringSubmatch(configuration)
		if versionMatch == nil || len(versionMatch) != 2 {
			continue
		}

		version := versionToInt(versionMatch[1])
		for _, storedVersion := range versions {
			if storedVersion == version {
				continue configurations
			}
		}

		versions = append(versions, version)
	}
	return versions
}

func versionToInt(input string) (version int) {
	re := regexp.MustCompile(`^(\d+).?(\d+)?.?(\d+)?`)
	numbersMatch := re.FindStringSubmatch(input)
	if numbersMatch == nil {
		return
	}

	for step, match := range numbersMatch[1:] {
		if number, err := strconv.Atoi(match); err == nil {
			version += int(number) * int(math.Pow10(4-(step*2)))
		}
	}
	return
}

func downloadWithRetries(ctx context.Context, sugar *zap.SugaredLogger, url string) (res *http.Response, err error) {
	if url == "" {
		return nil, errors.New("missing URL to download")
	}

	client := &http.Client{Timeout: 60 * time.Second}

	for retry := range []int{1, 2, 3} {
		sugar.Infof("Obtaining %s", url)

		var request *http.Request
		request, err = http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		request.Header.Set("Accept", "application/json")
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/70.0")
		request = request.WithContext(ctx)
		res, err = client.Do(request)
		if err == nil {
			break
		}

		time.Sleep(time.Duration(retry*retry) * time.Second)
	}
	return
}
