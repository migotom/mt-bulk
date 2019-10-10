package entities

// Command specifies single command, expected (or not) comamnd's result and optional sleep time that should be performed after command execution.
type Command struct {
	Body        string `toml:"body" yaml:"body"`
	Expect      string `toml:"expect" yaml:"expect"`
	MatchPrefix string `toml:"match_prefix" yaml:"match_prefix"`
	Match       string `toml:"match" yaml:"match"`
	SleepMs     int    `toml:"sleep_ms" yaml:"sleep_ms"`
}

func (c Command) String() string {
	return c.Body
}
