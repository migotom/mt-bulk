package entities

import "encoding/json"

// Command specifies single command, expected (or not) comamnd's result and optional sleep time that should be performed after command execution.
type Command struct {
	Body        string `toml:"body" yaml:"body" json:"body"`
	Expect      string `toml:"expect" yaml:"expect" json:"expect"`
	MatchPrefix string `toml:"match_prefix" yaml:"match_prefix" json:"match_prefix"`
	Match       string `toml:"match" yaml:"match" json:"match"`
	SleepMs     int    `toml:"sleep_ms" yaml:"sleep_ms" json:"sleep_ms"`
}

func (c Command) String() string {
	return c.Body
}

// CommandResult defines result of execution single command/operation.
type CommandResult struct {
	Body      string   `json:"body"`
	Responses []string `json:"responses"`
	Error     error    `json:"error"`
}

// MarshalJSON marshals CommandResult with error support.
func (r *CommandResult) MarshalJSON() ([]byte, error) {
	var err string
	if r.Error != nil {
		err = r.Error.Error()
	}

	type Copy CommandResult
	return json.Marshal(&struct {
		Error string `json:"error,omitempty"`
		*Copy
	}{
		Copy:  (*Copy)(r),
		Error: err,
	})
}
