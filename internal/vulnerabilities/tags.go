package vulnerabilities

// CVEURL is default CVE search API endpoint.
const CVEURL = "https://cve.circl.lu/api/search/mikrotik"

// RequiredKVDBVersion defines latest KV structure version.
const RequiredKVDBVersion = 1

const (
	kvTagCVE          = "CVE:"
	kvTagVersion      = "Version:"
	kvTagDBLastUpdate = "DB:LastUpdate"
	kvTagDBVersion    = "DB:Version"
)
