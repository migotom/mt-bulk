# mt-bulk

MT-bulk asynchronously and parallel sends using Mikrotik SSL API or SSH defined commands to list of devices provided by command line, loaded from file or database.

## Options

```
MT-bulk.

Usage:
  mt-bulk gen-certs [options]
  mt-bulk init-secure-api [options] [<hosts>...]
  mt-bulk change-password (--new=<newpass>) [options] [<hosts>...]  
  mt-bulk custom-api [--commands-file=<commands>] [options] [<hosts>...]  
  mt-bulk custom-ssh [--commands-file=<commands>] [options] [<hosts>...]  
  mt-bulk -h | --help
  mt-bulk --version

Options:
  -C <config-file>         Use configuration file, e.g. certs locations, ports, commands sequences, custom commands, etc...

  -s                       Be quiet and don't print commands and commands results to standard output
  -w <workers>             Number of parallel connections to run (default: 4)

  --skip-version-check     Skip checking for new version
  --skip-summary           Skip errors summary
  --exit-on-error          In case of any error stop executing commands

  --source-db              Load hosts using database configured by -C <config-file>
  --source-file=<file-in>  Load hosts from file <file-in>

  <hosts>...               List of space separated hosts in format IP[:PORT]
```

## Download

Current and historical releases of MT-bulk https://github.com/migotom/mt-bulk/releases

## Examples

```
mt-bulk gen-certs -C mtbulk.cfg
```
Create new CA, host and device certificates.

```
mt-bulk init-secure-api -C mtbulk.cfg 192.168.1.2 192.168.1.3:222 192.168.1.4:6654
```
Initialize 192.168.1.2, 192.168.1.3 and 192.168.1.4 (each device has running SSH on different port) with SSL API and certificates pointed by mtbulk.cfg (section [certificates_store]).

```
mt-bulk change-password -w 16 -C mtbulk.cfg --new=supersecret --source-db
```
Change password to *supersecret* on devices selected by SQL query using 16 workers. Connection details and query pointed by mtbulk.cfg (section [db])


## Operations

List of possible operations.

Syntax:
```
mt-bulk <operation-name> [additional parameters]
```

MT-bulk while connecting to devices is using list of passwords provided in configuration file, sections: [service.ssh] and [service.mikrotik_api], passwords are comma separated.

e.g.
```
[service]
[service.ssh]
port = "22"
password = "most_common_secret,alternative_secret,old_secret"
user = "admin"
```

### gen-certs

Generate and store device and host certificates required to establish secure connection using Mikrotik API. As default certificates are stored in certs/ folder. 
This operation may be proceeded once, MT-bulk will use certificates from certs/ to handle connections with each device.

### init-secure-api

Initialize Mikrotik device to use secure API with MT-bulk. Operation uploads to device certificate and enables api-ssl with given certificate.

### change-password

Change password to given new one with option `--new=<newpass>`
Important note: operation require SSL API already initialized.

### custom-api and custom-ssh

Send sequence of commands defined in configuration file:

Example:
```
[[custom-ssh.command]]
body = "/certificate print detail"
sleep = "1s"
match_prefix = "c"
match = '(?m)^\s+(\d+).*mtbulkdevice'
[[custom-ssh.command]]
body = "/certificate remove %{c1}"
sleep = "100ms"
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
- sleep: wait given time duration after executing command, required by some commands (e.g. `/system upgrade refresh`)
- expect: regexp used to verify that command's response match expected value
- match: regexp used to search value in command's output, using Go syntax https://github.com/google/re2/wiki/Syntax 
- match_prefix: for each match MT-bulk builds matcher using match_prefix and numbered capturing group, eg. %{prefix0}, %{prefix1} ...

More examples at `mt-bulk.example.cfg` and `example-commands\` folder.

## Configuration

### Format

Configuration file is written in TOML format (https://github.com/toml-lang/toml)

### Loading sequence 

- Application defaults
- System (/etc/mt-bulk/config.cfg, /Library/Application Support/MT-bulk/config.cfg)
- Home (~/.mt-bulk.cfg, ~/Library/Application Support/Uping/config.cfg)
- Command line -C option

### Hosts

Host can be specified using format:
- <ip>
- <ip>:<port>

This rule applies to hosts loaded using `--source-file` or provided directly by `[<hosts>...]`

## Troubleshooting

### SSH connections issues

- Verify host with running MT-bulk have access to Mikrotik device
- Verify username, password, host and port are valid, double check using eg. OpenSSH (MT-bulk doesn't require any additional tools or libraries, uses builtin in runtime SSH implementation)
- Some older RouterOS allows old and unsecure ciphers, SSH implementation builtinto MT-bulk will not establish connection using such ciphers, please upgrade your Mikrotik/RouterOS device
- Use strong-crypto by setting `/ip ssh set strong-crypto=yes` on RouterOS
- If nothing helps please provide log of establishing connection using ssh command `ssh -vvv <user>@<ip>:<port>` 

## Credits

Application was developed by Tomasz Kolaj and is licensed under Apache License Version 2.0.
Please reports bugs at https://github.com/migotom/mt-bulk/issues.
