package entities

import (
	"encoding/json"
)

// Result response from worker.
type Result struct {
	Job                   Job             `toml:"job" yaml:"job" json:"-"`
	Results               []CommandResult `toml:"results" yaml:"results" json:"results,omitempty"`
	DownloadURLs          []string        `json:"download_urls,omitempty"`
	AdditionalInformation []string        `json:"additional_information,omitempty"`
	Errors                []error         `toml:"errors" yaml:"errors" json:"errors,omitempty"`
}

// MarshalJSON marshals Result with error support.
func (r *Result) MarshalJSON() ([]byte, error) {
	var err []string
	if r.Errors != nil {
		for _, e := range r.Errors {
			if e != nil {
				err = append(err, e.Error())
			}
		}
	}

	type Copy Result
	return json.Marshal(&struct {
		Errors []string `json:"errors,omitempty"`
		*Copy
	}{
		Copy:   (*Copy)(r),
		Errors: err,
	})
}
