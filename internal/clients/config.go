package clients

// Clients represents list of supported clients types.
type Clients struct {
	SSH         Config `toml:"ssh"`
	MikrotikAPI Config `toml:"mikrotik_api"`
}

// Config of client.
type Config struct {
	VerifySleepMs int    `toml:"verify_check_sleep_ms"`
	Retries       int    `toml:"retries_count"`
	KeyStore     string `toml:"keys_store"`
	Pty           Pty    `toml:"pty"`

	DefaultPort     string `toml:"port"`
	DefaultUser     string `toml:"user"`
	DefaultPassword string `toml:"password"`
}

// Pty definition for SSH.
type Pty struct {
	Widht  int `toml:"width"`
	Height int `toml:"height"`
}

// NewConfig returns default config.
func NewConfig(port string) Config {
	return Config{
		VerifySleepMs: 1000,
		Retries:       2,
		Pty: Pty{
			Widht:  120,
			Height: 200,
		},
		DefaultPort: port,
	}
}
