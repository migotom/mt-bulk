package entities

// Command specifies single command, expected (or not) comamnd's result and optional sleep time that should be performed after command execution.
type Command struct {
	Body        string `toml:"body"`
	Expect      string `toml:"expect"`
	MatchPrefix string `toml:"match_prefix"`
	Match       string `toml:"match"`
	SleepMs     int    `toml:"sleep_ms"`
}

func (c Command) String() string {
	return c.Body
}
