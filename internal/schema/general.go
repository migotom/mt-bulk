package schema

import "time"

type ModeHandlerFunc func(*GeneralConfig, Host) error

// GeneralConfig main application configuration.
type GeneralConfig struct {
	IgnoreErrors      bool
	Verbose           bool
	SkipSummary       bool
	Workers           int   `toml:"workers"`
	VerifySleep       int   `toml:"verify_check_sleep"`
	Certs             Certs `toml:"certificates_store"`
	Service           map[string]Service
	ModeHandler       ModeHandlerFunc
	DB                DBConfig
	CustomSSHSequence CustomSequence `toml:"custom-ssh"`
	CustomAPISequence CustomSequence `toml:"custom-api"`
}

type Service struct {
	DefaultPort string `toml:"port"`
	DefaultUser string `toml:"user"`
	DefaultPass string `toml:"password"`
}
type Certs struct {
	Directory string
	Generate  bool
}

type CustomSequence struct {
	Command []Command
}

// Command specifies single command, expected (or not) comamnd's result and optional sleep time that should be performed after command execution.
type Command struct {
	Body         string `toml:"body"`
	SkipExpected bool   `toml:"skip_expected"`
	Expect       string `toml:"expect"`
	MatchPrefix  string `toml:"match_prefix"`
	Match        string `toml:"match"`

	Sleep  Duration
	Result string
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
