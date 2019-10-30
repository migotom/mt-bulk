# MT-bulk configuration

### Format

MT-bulk supports two formats of configuration files:

- TOML format (https://github.com/toml-lang/toml)
- YAML format (https://yaml.org/spec/)

Default configuration format since version 2.x is YAML.

### Main

| Property       | Default | Summary                                                       |
| -------------- | ------- | ------------------------------------------------------------- |
| `version`      | 2       | version of configuration file, MT-bulk 2.x requires version 2 |
| `verbose`      | true    | print commands' execution output                              |
| `skip_summary` | false   | skip summary of errors                                        |
| `service`      |         | section defining setup of service                             |
| `db`           |         | section defining setup of database connection                 |

### Service

| Property             | Default | Summary                                                  |
| -------------------- | ------- | -------------------------------------------------------- |
| `workers`            | 4       | number of parallel workers executing jobs                |
| `skip_version_check` | false   | do not check new mt-bulk version                         |
| `clients`            |         | section defining setup of all clients implementations    |
| `cve_urls`           |         | url list used to fetchvMikrotik's CVEs (can be empty)    |

### Clients

| Property       | Default | Summary                                           |
| -------------- | ------- | ------------------------------------------------- |
| `ssh`          |         | section defining setup of SSH client              |
| `mikrotik_api` |         | section defining setup of Mikrotik SSL API client |

### Client (SSH/Mikrotik API)

| Property                | Default    | Summary                                                                                                                                                                                   |
| ----------------------- | ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `verify_check_sleep_ms` | 1000       | sleep specified amount of milliseconds after command execution (it is required by some of Mikrotik commands before sending new command)                                                   |
| `retries`               | 3          | number of retries while establishing connection                                                                                                                                           |
| `port`                  | 22 or 8729 | default port used to establish connection (if not provided in host configuration)                                                                                                         |
| `password`              |            | list of passwords separated by comma to use while connecting to device (if not provided in host configuration)                                                                            |
| `user`                  |            | user name used to establish connection (if not provided in host configuration)                                                                                                            |
| `keys_store`            |            | location of folder with public/private keys (in case of SSH) or keys and certificates (in case of Mikrotik API) used to establish secure connection or used to authenticate by public key |
| `pty`                   |            | pty settings for SSH                                                                                                                                                                      |
### CVE URLs

| Property   | Default | Summary                                                      |
| ---------- | ------- | ------------------------------------------------------------ |
| `db_info`  |         | API endpoint used to fetch CVE repositories last updat       |
| `db`       |         | API endpoint used to fetch CVE database with Mikrotik issues |


### DB

| Property    | Default | Summary                                                    |
| ----------- | ------- | ---------------------------------------------------------- |
| `driver`    |         | database driver (currently ony `postgres` is supported)    |
| `params`    |         | connection parameters eg. `"postgres://user:pass@host/db"` |
| `id_server` |         | parameter passed by `get_devices` query                    |
| `queries`   |         | list of queries                                            |
