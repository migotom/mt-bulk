package schema

// DBConfig defines database connection settings.
type DBConfig struct {
	Driver     string
	Params     string
	Connection interface{}
	IDserver   int `toml:"id_server"`
	Queries    DBQueries
}

// DBQueries defines list of database queries.
type DBQueries struct {
	GetDevices   string `toml:"get_devices"`
	UpdateDevice string `toml:"update_device"`
}
