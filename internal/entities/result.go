package entities

import "encoding/json"

// Result response from worker.
type Result struct {
	Job          Job             `toml:"job" yaml:"job" json:"-"`
	Results      []CommandResult `toml:"results" yaml:"results" json:"results,omitempty"`
	DownloadURLs []string        `json:"download_urls,omitempty"`
	Error        error           `toml:"error" yaml:"error" json:"error,omitempty"`
}

// MarshalJSON marshals Result with error support.
func (r *Result) MarshalJSON() ([]byte, error) {
	var err string
	if r.Error != nil {
		err = r.Error.Error()
	}

	type Copy Result
	return json.Marshal(&struct {
		Error string `json:"error,omitempty"`
		*Copy
	}{
		Copy:  (*Copy)(r),
		Error: err,
	})
}
