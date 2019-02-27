# mt-bulk

MT-bulk asynchronously and parallel sends using Mikrotik SSL API or SSH defined operations to list of devices provided by command line, loaded from file or database.

## Options

```
MT-bulk.

Usage:
  mt-bulk gen-certs [options]
  mt-bulk init-secure-api [options] [<hosts>...]
  mt-bulk change-password (--new=<newpass>) [options] [<hosts>...]  
  mt-bulk custom-api [--custom-args=<custom-args>] [options] [<hosts>...]  
  mt-bulk custom-ssh [--custom-args=<custom-args>] [options] [<hosts>...]  
  mt-bulk -h | --help
  mt-bulk --version

Options:
  -C <config-file>         Use configuration file, eg. certs locations, ports, commands sequences, custom commands, etc...

  -s                       Be quiet and don't print commands and commands results to standard output
  -w <workers>             Number of paralell connections to run (default: 4)

  --skip-summary           Skip errors summary
  --exit-on-error          In case of any error stop executing commands

  --source-db              Load hosts using database configured by -C <config-file>
  --source-file=<file-in>  Load hosts from file <file-in>

  <hosts>...               List of space separated hosts in format IP[:PORT]
```

## Operations

List of possible operations.

Syntax:
```
mt-bulk <operation-name> [additional parameters]
```

MT-bulk while connecting to devices is using list of passwords provided in configuration file, sections: [service.ssh] and [service.mikrotik_api], passwords are comma separated.

eg.
```
[service]
[service.ssh]
port = "22"
password = "most_common_secret,alternative_secret,old_secret"
user = "admin"
```

### gen-certs

Generate and store device and host certificates required to establish secure connection using Mikrotik API. As default certificates are stored in certs/ folder. 
This operation may be proceed once, mt-bulk will use certificates from certs/ to handle connections with each device.

### init-secure-api

Initialize Mikrotik device to use secure API with mt-bulk. Operation uploads to device certificate and enables api-ssl with given certificate.

### change-password

Change passwor to given new one with option --new=< newpass >

### custom-api and custom-ssh

Send sequence of commands defined in configuration file.

Sample:
```
[[custom-ssh.command]]
body = "/certificate print detail"
match_prefix = "c"
match = '(?m)^\s+(\d+).*mtbulkdevice'
[[custom-ssh.command]]
body = "/certificate remove %{c1}"
```

Command's options:
- body: command with parameters, allowed to use regex matches in format %{[prefix][number of numbered capturing group]}
- match: regexp used to search value in command's output, using Go syntax https://github.com/google/re2/wiki/Syntax 
- match_prefix: for each match mt-bulk builds matcher using match_prefix and numbered capturing group, eg. %{prefix0}, %{prefix1} ...
