# MT-bulk REST API configuration

### Format

MT-bulk supports two formats of configuration files:
* TOML format (https://github.com/toml-lang/toml)
* YAML format (https://yaml.org/spec/)

Default configuration format since version 2.x is YAML.

### Main

Property | Default | Summary
----|---------|------------
`version`|2| version of configuration file, MT-bulk 2.x requires version 2
`listen`| | address to listen to eg. `":8080"`
`root_directory`|| root directory used to store files accessible by HTTP request or transferred to/from Mikrotik devices using SFTP
`keys_store`|| directory containing private/public key used to establish HTTPS session
`token_secret`|| secret used to sign tokens
`authenticate`|| section defining authentication/authorization rules

Rest of sections have identical configuration like command line version of MT-bulk.