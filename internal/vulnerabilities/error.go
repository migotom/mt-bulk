package vulnerabilities

import (
	"fmt"
	"strings"
)

// VulnerabilityError is error containing all found vulnerabilities.
type VulnerabilityError struct {
	Vulnerabilities []CVE
	err             error
}

func (e VulnerabilityError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}

	if len(e.Vulnerabilities) == 0 {
		return ""
	}
	var err strings.Builder

	err.WriteString("vulnerabilities found: ")
	for _, cve := range e.Vulnerabilities {
		err.WriteString(fmt.Sprintf("%s (%.1f), ", cve.ID, cve.CVSS))
	}
	return strings.TrimRight(err.String(), ", ")
}

// Details returns details about found vulnerabilities.
func (e VulnerabilityError) Details() []string {
	details := make([]string, 0, len(e.Vulnerabilities))

	for _, cve := range e.Vulnerabilities {
		details = append(details, cve.String())
	}
	return details
}
