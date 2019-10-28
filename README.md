# mt-bulk

MT-bulk is a toolset to help manage multiple Mikrotik/RouterOS devices by sending in parallel predefined or custom commands to multiple devices at once using Mikrotik SSL API, SSH and SFTP.

Version 2.x introduces a major breaking changes, please read [Version 2 breaking changes](#Version-2-breaking-changes) section of this readme before upgrading.

MT-bulk toolset contains two tools:

### Mt-bulk

CLI tool that process devices list and commands provided by command line arguments, loaded from file or external SQL database. Commands are distributed to several internal workers and processed parallel.

### MT-bulk REST API gateway

Simple REST API server that process HTTPS POST requests with specified pair of commands and hosts to asynchronously execute on.

### Contents

- [Command's options](#Options)
- [Download](#Download)
- [List of possible operations/modes](#Operations)
- [Configuration description](#Configurations)
- [Quick Overview / Tutorial](#Quick-Overview-/-Tutorial)
- [Troubleshooting](#Troubleshooting)

## Options

### MT-bulk

```
MT-bulk.

Usage:
  mt-bulk gen-api-certs [options]
  mt-bulk gen-ssh-keys [options]
  mt-bulk init-secure-api [options] [<hosts>...]
  mt-bulk init-publickey-ssh [options] [<hosts>...]
  mt-bulk change-password (--new=<newpass>) [--user=<login>] [options] [<hosts>...]
  mt-bulk system-backup (--name=<name>) (--backup-store=<backups>) [options] [<hosts>...]
  mt-bulk security-audit [options] [<hosts>...]
  mt-bulk sftp <source> <target> [options] [<hosts>...]

  mt-bulk custom-api [--commands-file=<commands>] [options] [<hosts>...]
  mt-bulk custom-ssh [--commands-file=<commands>] [options] [<hosts>...]
  mt-bulk -h | --help
  mt-bulk --version

Options:
  -C <config-file>         Use configuration file, e.g. keys/certs locations, ports, commands sequences, custom commands, etc...
  --source-db              Load hosts using database configured by -C <config-file>
  --source-file=<file-in>  Load hosts from file <file-in>

  <hosts>...               List of space separated hosts in format IP[:PORT]
```

### MT-bulk REST API gateway

```
MT-bulk REST API gateway.

Usage:
  mt-bulk-rest-gw [options]
  mt-bulk-rest-gw gen-https-certs [options]
  mt-bulk-rest-gw -h | --help
  mt-bulk-rest-gw --version

Options:
  -C <config-file>         Use configuration file
```

## Download

Current and historical releases of MT-bulk and MT-bulk-rest-api at https://github.com/migotom/mt-bulk/releases

## Operations

List of possible operations to execute by CLI and REST API:

- [Generate Mikrotik API SSL certificate](./docs/operations.md#Generate-Mikrotik-API-SSL-certificates)
- [Generate SSH RSA Private/Public keys](./docs/operations.md#Generate-SSH-RSA-Private/Public-keys)
- [Initialize device to use Mikrotik SSL API](./docs/operations.md#Initialize-device-to-use-Mikrotik-SSL-API)
- [Initialize device to use Public key SSH authentication](./docs/operations.md#Initialize-device-to-use-Public-key-SSH-authentication)
- [Change user's password](./docs/operations.md#Change-user's-password)
- [System backup](/docs/operations.md#System-backup)
- [SFTP](/docs/operations.md#SFTP)
- [Security audit](/docs/operations.md#Security-audit)
- [Execute sequence of custom commands](./docs/operations.md#Execute-sequence-of-custom-commands)

## Configurations

### Format

MT-bulk supports two formats of configuration files:

- TOML format (https://github.com/toml-lang/toml)
- YAML format (https://yaml.org/spec/)

Default configuration format since version 2.x is YAML.

### Configurations loading sequence

- Application defaults
- System (`/etc/mt-bulk/config.yml`, `/Library/Application Support/MT-bulk/config.yml`)
- Home (`~/.mt-bulk.yml`, `~/Library/Application Support/MT-bulk/config.yml`)
- Command line `-C` option

### Hosts

Host can be specified using:

- simple text format:

  - `ip`
  - `ip:port`
  - `foo.bar.com:port`

- using YAML:

```yaml
host:
  - ip: "192.168.1.1"
    password: "secret"
  - ip: "192.168.1.2:22"
    user: "john"
    password: "secret"
```

### Detailed configuration descriptions

- [MT-bulk command line tool
  ](./docs/configuration-mt-bulk.md#MT-bulk-configuration)
- [MT-bulk REST API](./docs/configuration-mt-bulk-rest-api.md#MT-bulk-REST-API-configuration)

## Quick Overview / Tutorial

Detailed examples and sample configurations at [examples/](examples/) directory of project.

### MT-bulk

```bash
mt-bulk gen-api-certs -C examples/configurations/mt-bulk.example.yml
```

Create new CA, host and device certificates.

```bash
mt-bulk init-secure-api -C examples/configurations/mt-bulk.example.yml 192.168.1.2 192.168.1.3:222 192.168.1.4:6654
```

Initialize 192.168.1.2, 192.168.1.3 and 192.168.1.4 (in this example each device has running SSH on different port) with user, SSL API and certificates pointed by mtbulk.cfg (section [service.clients.mikrotik_api.keys_store]).

```bash
mt-bulk change-password -w 16 -C examples/configurations/mt-bulk.example.yml --new=supersecret --user=admin --source-db
```

Change admin's password to _supersecret_ for admin user on devices selected by SQL query using 16 workers. Connection details and query pointed by mtbulk.cfg (section [db])

### MT-bulk REST API gateway

```bash
mt-bulk-rest-api gen-https-certs -C examples/configurations/mt-bulk-rest-api.example.yml
```

Generates self signed SSL certificates and starts REST API daemon.

```bash
mt-bulk-rest-api -C examples/configurations/mt-bulk-rest-api.example.yml
```

Request authentication token:

```bash
$ curl -k -H "Content-Type: application/json" -d '{"key":"abc"}' -X POST https://localhost:8080/authenticate

{"token":"yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy.zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz.xxxxxxxxxxxxxxxxxxx"}
```

Request execute custom command:

```bash
$ curl -k -H "Authorization: some.token.value" -H "Content-Type: application/json" -d '{"host":{"ip":"10.0.0.1","user":"admin","password":"secret"},"kind":"CustomSSH","commands":[{"body":"/user print","expect":"LAST"}]}}' -X POST https://localhost:8080/job

{"results":[{"body":"/\u003cmt-bulk\u003eestablish connection","responses":["/\u003cmt-bulk\u003eestablish connection"," --\u003e attempt #0, password #0, job #bmnb10sllhcjk1tifn2g"]},{"body":"/user print","responses":["/user print\nFlags: X - disabled \n #   NAME                                 GROUP                                 ADDRESS            LAST-LOGGED-IN      \n 0   ;;; system default user\n     admin                                full                                                     jul/19/2019 07:38:54\n"]}]}

```

#### Endpoints

- POST https://localhost:8080/authenticate \
  Authenticate and obtain auth token. `"key"` is one of access keys defined in configuration [`authenticate.key`], each `"key"` can have list of regexp rules defining list of allowed device IP addresses to use in requests.

```json
{
  "key": "abc"
}
```

- POST https://localhost:8080/job \
  MT-bulk API request. Run and execute specified job with optional additional commands on specified host. Each request must have valid token as `Authorization` header field. [List of possible operations](./docs/operations.md)

```json
{
  "host": {
    "ip": "10.0.0.1",
    "user": "admin",
    "password": "secret"
  },
  "kind": "CustomSSH",
  "commands": [{ "body": "/user print", "expect": "LAST-LOGGED-IN" }]
}
```

- GET https://localhost:8080/{root_directory}/{path_to_file} \
  Download file (eg. uploaded earlier by `SystemBackup` operation to MT-bulk `root_directory` defined in configuration). Each request must have valid token as `Authorization` header field.

- POST https://localhost:8080/upload \
  Upload a file as multipart/form-data request with field `file`. Uploaded file is accessible in `root_directory` to operations like `SFTP` or `SystemBackup`. Each request must have valid token as `Authorization` header field.

## Troubleshooting

### SSH connections issues

- Verify host with running MT-bulk have access to Mikrotik device
- Verify username, password, host and port are valid, double check using eg. OpenSSH (MT-bulk doesn't require any additional tools or libraries, uses builtin in runtime SSH implementation)
- Some older RouterOS allows old and insecure ciphers, SSH implementation builtin MT-bulk will not establish connection using such ciphers, please upgrade your Mikrotik/RouterOS device
- Use strong-crypto by setting `/ip ssh set strong-crypto=yes` on RouterOS
- If nothing helps please provide log of establishing connection using ssh command `ssh -vvv <user>@<ip>:<port>`

## Version 2 breaking changes

### Command line options

CLI in version 2.x is simplified, all switches and configuring options are moved into configuration file, CLI is used to specify operation mode, configuration file and source of hosts to parse.

### Configuration file

Configuration structure is rewritten and divided into few sections. Some of options changed as well (eg. `verify_check_sleep` into `verify_check_sleep_ms`). Please compare your current configuration file with attached `mt-bulk.example.cfg`.
To let know to MT-bulk that configuration file is compatible with version 2.x new entry in config file was added: `version = 2`.

## Credits

Application was developed by Tomasz Kolaj and is licensed under Apache License Version 2.0.
Please reports bugs at https://github.com/migotom/mt-bulk/issues.
