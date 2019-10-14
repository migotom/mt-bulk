# mt-bulk

MT-bulk is a toolset to help manage multiple Mikrotik/RouterOS devices by sending predefined or custom commands using Mikrotik SSL API, SSH and SFTP.  

Version 2.x introduces a major breaking changes, please read [Version 2 breaking changes](#Version-2-breaking-changes) section of this readme before upgrading.

MT-bulk toolset contains two tools:

## Mt-bulk

CLI tool that process devices list and commands provided by command line arguments, loaded from file or external SQL database. Commands are distributed to several internal workers and processed parallel.

### MT-bulk REST API gateway

REST API daemon that process HTTPS POST requests with specified pair of commands and hosts to asynchronously execute on.


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

## Examples

### MT-bulk

```
mt-bulk gen-api-certs -C mtbulk.cfg
```
Create new CA, host and device certificates.

```
mt-bulk init-secure-api -C mtbulk.cfg 192.168.1.2 192.168.1.3:222 192.168.1.4:6654
```
Initialize 192.168.1.2, 192.168.1.3 and 192.168.1.4 (in this example each device has running SSH on different port) with user, SSL API and certificates pointed by mtbulk.cfg (section [service.clients.mikrotik_api.keys_store]).

```
mt-bulk change-password -w 16 -C mtbulk.cfg --new=supersecret --user=admin --source-db
```
Change admin's password to *supersecret* on devices selected by SQL query using 16 workers. Connection details and query pointed by mtbulk.cfg (section [db])

### MT-bulk REST API gateway

```
mt-bulk-rest-api gen-https-certs -C mtbulk-rest-api.cfg
```
Generates self signed SSL certificates and starts REST API daemon.

```
mt-bulk-rest-api -C mtbulk-rest-api.cfg
```
Stars REST API daemon


## Operations

List of possible operations.

Syntax:
```
mt-bulk <operation-name> [additional parameters]
```

MT-bulk while connecting to devices is using SSH public key or list of passwords provided in configuration file, sections: [service.clients.ssh] and [service.clients.mikrotik_api], passwords are comma separated.

e.g.
```
[service]
[service.clients.ssh]
port = "22"
password = "most_common_secret,alternative_secret,old_secret"
user = "admin"
keys_store  = "keys/ssh"
```

### gen-api-certs

Generate and store device and host certificates required to establish secure connection using Mikrotik API. 
This operation may be proceeded once, MT-bulk will use certificates from [service.clients.mikrotik_api.keys_store] to handle connections with each device.

### gen-ssh-keys

Generate and store private and public SSH RSA keys that may be used to establish secure connection using SSH without password. 
This operation may be proceeded once, MT-bulk will use keys from [service.clients.ssh.keys_store] to handle connections with each device.

Important note. Password will not work once public key authentication enabled on RouterOS.

### init-secure-api

Initialize Mikrotik device to use SSL API with MT-bulk. Operation uploads to device certificate and enables api-ssl with given certificate.

### init-publickey-ssh

Initialize Mikrotik device to use SSH public key to authenticate with MT-bulk. Operation uploads to device public key and enables it for given user.

### change-password

Change password to given new one with option `--new=<newpass>` and optionally `--user=<user>`
Important note: operation require SSL API already initialized.

### custom-api and custom-ssh

Send sequence of commands defined in configuration file:

Example:
```
[[custom-ssh.command]]
body = "/certificate print detail"
sleep_ms = 1000
match_prefix = "c"
match = '(?m)^\s+(\d+).*mtbulkdevice'
[[custom-ssh.command]]
body = "/certificate remove %{c1}"
sleep_ms = 100
[[custom-ssh.command]]
body = "/system upgrade upgrade-package-source add address=10.0.0.1 user=test"
expect = "password:"
[[custom-ssh.command]]
body = "my_password"
```

Sequences may be also defined in separate files and provided to MT-bulk by `--commands-file=<file name>`:

Example SSH command:
```
[[command]]
body = "/user print"
```

Command's options:
- body: command with parameters, allowed to use regex matches in format %{[prefix][number of numbered capturing group]}
- sleep_ms: wait given time duration after executing command, required by some commands (e.g. `/system upgrade refresh`)
- expect: regexp used to verify that command's response match expected value
- match: regexp used to search value in command's output, using Go syntax https://github.com/google/re2/wiki/Syntax 
- match_prefix: for each match MT-bulk builds matcher using match_prefix and numbered capturing group, eg. %{prefix0}, %{prefix1} ...

More examples at `mt-bulk.example.cfg` and `example-commands\` folder.

## Configuration

### Format

Configuration file is written in TOML format (https://github.com/toml-lang/toml)

### Loading sequence 

- Application defaults
- System (`/etc/mt-bulk/config.cfg`, `/Library/Application Support/MT-bulk/config.cfg`)
- Home (`~/.mt-bulk.cfg`, `~/Library/Application Support/Uping/config.cfg`)
- Command line `-C` option

### Hosts

Host can be specified using format:
- `ip`
- `ip:port`
- `foo.bar.com:port`

This rule applies to hosts loaded using `--source-file` or provided directly by `[<hosts>...]`

## Troubleshooting

### SSH connections issues

- Verify host with running MT-bulk have access to Mikrotik device
- Verify username, password, host and port are valid, double check using eg. OpenSSH (MT-bulk doesn't require any additional tools or libraries, uses builtin in runtime SSH implementation)
- Some older RouterOS allows old and unsecure ciphers, SSH implementation builtinto MT-bulk will not establish connection using such ciphers, please upgrade your Mikrotik/RouterOS device
- Use strong-crypto by setting `/ip ssh set strong-crypto=yes` on RouterOS
- If nothing helps please provide log of establishing connection using ssh command `ssh -vvv <user>@<ip>:<port>` 

## Version 2 breaking changes

### Command line options

CLI in version 2.x is simplified, all switches and configuring options are moved into configuration file, CLI is used to specify operation mode, configuration file and source of hosts to parse.

### Configuration file

Configuration structure is rewrited and divided into few sections. Some of options changed as well (eg. `verify_check_sleep` into `verify_check_sleep_ms`). Please compare your current configuratio file with attached `mt-bulk.example.cfg`. 
To let know to MT-bulk that configuration file is compatible with version 2.x new entry in config file was added: `version = 2`.

## Credits

Application was developed by Tomasz Kolaj and is licensed under Apache License Version 2.0.
Please reports bugs at https://github.com/migotom/mt-bulk/issues.
