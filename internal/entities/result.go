package entities

// Result response from worker.
type Result struct {
	Host      Host     `toml:"host" yaml:"host"`
	Job       Job      `toml:"job" yaml:"job"`
	Responses []string `toml:"responses" yaml:"reponses"`
	Error     error    `toml:"error" yaml:"error"`
}
