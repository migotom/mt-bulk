package vulnerabilities

// RequiredKVDBVersion defines latest KV structure version.
const RequiredKVDBVersion = 1

const (
	kvTagCVE          = "CVE:"
	kvTagVersion      = "Version:"
	kvTagDBLastUpdate = "DB:LastUpdate"
	kvTagDBVersion    = "DB:Version"
	kvTagDBCVEdbInfo  = "DB:CVE:DBInfo"
)

// CVEURL is default CVE search API endpoint.
const CVEURL = "https://cve.circl.lu/api/search/mikrotik"

// CVEURLDBInfo is default CVE database timestamps information API endpoint.
const CVEURLDBInfo = "https://cve.circl.lu/api/dbInfo"

// CVEURLfallback is fallback CVE search database location.
const CVEURLfallback = "https://raw.githubusercontent.com/migotom/mt-bulk/master/utils/cves/cve_circl_mikrotik.json"

// CVEURLfallbackDBInfo is fallback CVE database timestamps location.
const CVEURLfallbackDBInfo = "https://raw.githubusercontent.com/migotom/mt-bulk/master/utils/cves/cve_circl_dbInfo.json"
