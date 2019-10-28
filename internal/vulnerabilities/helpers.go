package vulnerabilities

import (
	"math"
	"regexp"
	"strconv"
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
