# MT-bulk operations

List of possible operations to execute by CLI and REST API tools.

Each of operation have two sections, example syntax to use as CLI util `mt-bulk` and example of REST API request to use with `mt-bulk-rest-api` daemon.

**List of operations**:

- [Generate Mikrotik API SSL certificate](#Generate-Mikrotik-API-SSL-certificates)
- [Generate SSH RSA Private/Public keys](#Generate-SSH-RSA-Private/Public-keys)
- [Initialize device to use Mikrotik SSL API](#Initialize-device-to-use-Mikrotik-SSL-API)
- [Initialize device to use Public key SSH authentication](#Initialize-device-to-use-Public-key-SSH-authentication)
- [Change user's password](#Change-user's-password)
- [System backup](#System-backup)
- [SFTP](#SFTP)
- [Scan for CVEs and security audit](#Security-audit)
- [Execute sequence of custom commands](#Execute-sequence-of-custom-commands)

## Generate Mikrotik API SSL certificates

Generate and store CA, device and host certificates required to establish secure connection using Mikrotik SSL API.
This operation may be proceeded once, MT-bulk will use certificates from [`service.clients.mikrotik_api.keys_store`] to handle connections with each device.

### CLI

```bash
mt-bulk gen-api-certs -C your.configuration.file.yml 10.0.0.1
```

### REST API request

Operation is not allowed to call using REST requests.

## Generate SSH RSA Private/Public keys

Generate and store private and public SSH RSA keys that may be used to establish secure connection using SSH without password.
This operation may be proceeded once, MT-bulk will use keys from [service.clients.ssh.keys_store] to handle connections with each device.

**Important note**. Password will not work once public key authentication enabled on RouterOS.

### CLI

```bash
mt-bulk gen-ssh-keys -C your.configuration.file.yml 10.0.0.1
```

### REST API request

Operation is not allowed to call using REST requests.

## Initialize device to use Mikrotik SSL API

Initialize Mikrotik device to use SSL API with MT-bulk. Operation uploads to device certificate and enables Mikrotik SSL API with given certificate.

### CLI

Using command line tool its important to remember that location of certificates store is pointed by provided configuration file [`service.clients.mikrotik_api.keys_store`].

```bash
mt-bulk init-secure-api -C your.configuration.file.yml 10.0.0.1
```

### REST API request

```json
{
  "host": {
    "ip": "10.0.0.1",
    "user": "admin",
    "password": "secret"
  },
  "kind": "InitSecureAPI",
  "data": {
    "keys_directory": "certs/api"
  }
}
```

## Initialize device to use Public key SSH authentication

Initialize Mikrotik device to use SSH RSA public key to authenticate with MT-bulk. Operation uploads to device public key (`id_rsa.pub`) and enables it for given user.

**Important note**. Password will not work once public key authentication enabled on RouterOS.

### CLI

Using command line tool its important to remember that location of certificates store is pointed by provided configuration file [`service.clients.ssh.keys_store`].

```bash
mt-bulk init-publickey-ssh -C your.configuration.file.yml 10.0.0.1
```

### REST API request

```json
{
  "host": {
    "ip": "10.0.0.1",
    "user": "admin",
    "password": "secret"
  },
  "kind": "InitPublicKeySSH",
  "data": {
    "keys_directory": "certs/ssh"
  }
}
```

## Change user's password

Change password to given new one with option `--new=<newpass>` and optionally `--user=<user>`

**Important note**. Operation requires SSL API already initialized.

### CLI

```bash
mt-bulk change-password --new=newsecret --user=admin -C your.configuration.file.yml 10.0.0.1
```

### REST API request

```json
{
  "host": {
    "ip": "10.0.0.1",
    "user": "admin",
    "password": "secret"
  },
  "kind": "ChangePassword",
  "data": {
    "user": "admin",
    "new_password": "newsecret"
  }
}
```

## System backup

Do backup of a system with option `--new=<name>` defining name of backup and `--backup-store=<backups>` as a location where to store backup.

### CLI

```bash
mt-bulk system-backup --name=backup --backup-store=backups/ -C your.configuration.file.yml 10.0.0.1
```

### REST API request

```json
{
  "host": {
    "ip": "10.0.0.1",
    "user": "admin",
    "password": "secret"
  },
  "kind": "SystemBackup",
  "data": {
    "name": "backup",
    "backups_store": "backups/"
  }
}
```

## SFTP

Transfer files to and from devices using SFTP protocol.

### CLI

```bash
mt-bulk sftp sftp://file_on_mikrotik.txt local_folder/file.txt -C your.configuration.file.yml 10.0.0.1
```

### REST API request

```json
{
  "host": {
    "ip": "10.0.0.1",
    "user": "admin",
    "password": "secret"
  },
  "kind": "SFTP",
  "data": {
    "source": "sftp://file_on_mikrotik.txt",
    "target": "local_folder/file.txt"
  }
}
```

## Security audit

Check device for any known vulnerabilities by searching CVE databases for particular Mikrotik version and using SSH look on device itself for known non-secure settings turned on.

### CLI

```bash
mt-bulk security-audit -C your.configuration.file.yml 10.0.0.1
```

### REST API request

```json
{
  "host": {
    "ip": "10.0.0.1",
    "user": "admin",
    "password": "secret"
  },
  "kind": "SecurityAudit"
}
```

**Important note**

Keep in mind that first security audit can take while as it tries to connect to public CVE database and perform quite long set of checking options commands on device itself (if public CVE search engine is not available at the moment `mt-bulk` will try to fetch last known and saved at github repository mirror).
Once CVEs database is downloaded each next _Security audit_ operation should perform much faster as `mt-bulk` caches list of known issues up to 24h.

## Execute sequence of custom commands

custom-api and custom-ssh

Sends sequence of custom commands, including optional options like verification of expected result, pattern matching and picking part of output one command to use by another one.

Command's options:

- body: command with parameters, allowed to use regex matches in format %{[prefix][number of numbered capturing group]}
- sleep_ms: wait given time duration after executing command, required by some commands (e.g. `/system upgrade refresh`)
- expect: regexp used to verify that command's response match expected value
- match: regexp used to search value in command's output, using Go syntax https://github.com/google/re2/wiki/Syntax
- match_prefix: for each match MT-bulk builds matcher using match_prefix and numbered capturing group, eg. %{prefix0}, %{prefix1} ...

### CLI

List of commands can be defined in main configuration file [`custom-ssh.command`] or [`custom-api.command`] sections.

Commands' sequences may be also defined in separate files and provided to MT-bulk by `--commands-file=<file name>`

Example configuration:

```yaml
custom-ssh:
  command:
    - body: "/certificate print detail"
      sleep_ms: 1000
      match_prefix: "c"
      match: "(?m)^\s+(\d+).*mtbulkdevice"
    - body: "/certificate remove %{c1}"
      sleep_ms: 100
    - body: "/system upgrade upgrade-package-source add address=10.0.0.1 user=test"
      expect: "password:"
    - body: "my-secret-password"
```

```bash
mt-bulk custom-ssh -C your.configuration.file.yml 10.0.0.1
```

### REST API request

Allowed `CustomSSH` and `CustomAPI` kind of job.

```json
{
  "host": {
    "ip": "10.0.0.1",
    "user": "admin",
    "password": "secret"
  },
  "kind": "CustomSSH",
  "commands": [
    {
      "body": "/user print",
      "expect": "NAME"
    },
    {
      "body": "/system upgrade upgrade-package-source add address=10.0.0.1 user=test",
      "expect": "password:"
    },
    {
      "body": "my-secret-password"
    }
  ]
}
```
