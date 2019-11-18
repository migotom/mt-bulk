package clients

// Clients represents list of supported clients types.
type Clients struct {
	SSH         Config `toml:"ssh" yaml:"ssh"`
	MikrotikAPI Config `toml:"mikrotik_api" yaml:"mikrotik_api"`
}

// Config of client.
type Config struct {
	VerifySleepMs int    `toml:"verify_check_sleep_ms" yaml:"verify_check_sleep_ms"`
	Retries       int    `toml:"retries" yaml:"retries"`
	KeyStore      string `toml:"keys_store" yaml:"keys_store"`
	Pty           Pty    `toml:"pty" yaml:"pty"`

	DefaultPort     string `toml:"port" yaml:"port"`
	DefaultUser     string `toml:"user" yaml:"user"`
	DefaultPassword string `toml:"password" yaml:"password"`
}

// Pty definition for SSH.
type Pty struct {
	Width  int `toml:"width" yaml:"width"`
	Height int `toml:"height" yaml:"height"`
}

// NewConfig returns default config.
func NewConfig(port string) Config {
	return Config{
		VerifySleepMs: 1000,
		Retries:       2,
		Pty: Pty{
			Width:  120,
			Height: 200,
		},
		DefaultPort: port,
	}
}
