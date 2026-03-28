# pwcli

Toolbox for validating, storing and querying encrypted passwords

![CI](https://github.com/tommi2day/pwcli/actions/workflows/main.yml/badge.svg)
[![codecov](https://codecov.io/gh/Tommi2Day/pwcli/branch/main/graph/badge.svg?token=3EBK75VLC8)](https://codecov.io/gh/Tommi2Day/pwcli)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tommi2day/pwcli)

## Contents

- [Features](#features)
- [Local Password Store](#local-password-store)
- [gopass Store](#gopass-store)
- [Password Profiles](#password-profiles)
- [Command Reference](#command-reference)
- [Examples](#examples)

## Features

This tool contains a collection of often-used solutions for:

- Retrieving passwords from encrypted files or password stores with different methods
  - RSA keys (default)
    - Go crypto functions
    - OpenSSL-compatible format (default)
  - KMS keys (Amazon KMS)
  - HashiCorp Vault
  - gopass-compatible password store (age or GPG encryption)
  - Plain (unencrypted) files
  - Base64-encoded files

- Generating and checking secure passwords with password profiles

- Generating hashes with common methods and verifying a password against them
  - MD5
  - SSHA (e.g. for LDAP passwords)
  - Argon2
  - BCrypt (e.g. for htpasswd)
  - HTTP Basic Auth format
  - SCRAM (e.g. for PostgreSQL)

- Generating RSA, ECDSA, age and GPG key pairs (optionally passphrase-protected)

- Generating TOTP codes

- LDAP operations
  - Set LDAP password
  - Set SSH public key field
  - Retrieve entry attributes
  - Look up group membership of a user
  - Look up users by group

---

## Local Password Store

You may have different password stores using different formats, keys and directories.
The identifier is the `application` name.

### Config files

This tool may use config files via viper.  The default filename is `<application>.yaml` with
a fallback to `pwcli.yaml`.  It will search:

- the current directory
- `$HOME/.pwcli`
- `$HOME/etc/`
- `/etc/pwcli`

You may supply an explicit path with `--config`.

The following saves a default config file `get_password.yaml` for the application `get_password`
in the current directory, pointing `datadir` and `keydir` at `$HOME/.pwcli` and using the
RSA-in-Go file format:

````shell
pwcli config save -a get_password -D $HOME/.pwcli -K $HOME/.pwcli -m go
````

Make sure the directories exist:

````shell
mkdir $HOME/.pwcli
````

### Key generation

`pwcli genkey` supports RSA, ECDSA, age and GPG key pairs.  Use `--type` to select the
key type (default `rsa`).  Use `--keypass` to protect the private key with a
passphrase — for age this creates an age scrypt-encrypted key file; for
RSA/ECDSA it AES-256 encrypts the PEM block; for GPG it locks the secret key.

Generate an RSA key pair using the config above:

````shell
pwcli genkey -a get_password
````

Generate a passphrase-protected age key pair:

````shell
pwcli genkey -a get_password --type age --keypass mysecret
````

### Password store file

When not using a third-party store (Vault, gopass), the local password store is built from
a plaintext file encrypted via the configured method.

The plaintext file uses colon-delimited `system:user:password` lines.  The special system
name `!default` matches any system:

```
# default match for this user on every system
!default:testuser:default
# exact match — takes precedence over !default
test:testuser:testpass
```

Name the file `<app>.plain` in `datadir`, or specify it explicitly with `--plaintext`:

````shell
pwcli encrypt -a get_password --plaintext test.plain
````

### Querying passwords

Search for system and user is case-insensitive by default (except for `vault` and `gopass`
methods).  Use `--case-sensitive` to enforce exact matching for all methods.

````shell
pwcli get -a get_password -u testuser -d test
testpass
````

When the private key is passphrase-protected and `--keypass` is not supplied, `get`,
`decrypt` and `sign` will prompt interactively as a last resort:

````shell
pwcli get -a get_password -u testuser -d test
Key passphrase: ••••••••
testpass
````

Use `--no-prompt` to suppress all interactive prompts and return an error instead — useful
in CI/batch pipelines:

````shell
pwcli --no-prompt get -a get_password -u testuser -d test
Error: …
````

---

## gopass Store

`pwcli` can read from and write to a gopass-compatible password store (age or GPG
encryption).  Use `--method gopass` with `get`, or the dedicated `gopass` subcommand for
full store management.

### Quick start

````shell
# Initialise a new store with a fresh age key
pwcli gopass identity create mykey --add-recipient -S /path/to/store

# Write a secret
pwcli gopass write db/prod/password -S /path/to/store --content "s3cr3t"

# Read it back (identity is auto-detected)
pwcli gopass read db/prod/password -S /path/to/store
s3cr3t
````

### Passphrase-protected age identity

````shell
# Create an encrypted identity
pwcli gopass identity create mykey --passphrase mysecret --add-recipient -S /path/to/store

# Read a secret — prompted for passphrase when --keypass is omitted
pwcli gopass read db/prod/password -S /path/to/store
Age identity passphrase: ••••••••
s3cr3t

# Supply passphrase non-interactively
pwcli gopass read db/prod/password -S /path/to/store --keypass mysecret

# Or via environment variable
GOPASS_AGE_PASSWORD=mysecret pwcli gopass read db/prod/password -S /path/to/store
````

### Use gopass store as method for get

````shell
pwcli get --method gopass --path db/prod --entry password
````

---

## Password Profiles

Password profile sets define character-class rules for generation and validation.
Several sets are predefined:

````yaml
default:
  profile:
    length: 16
    upper: 1
    lower: 1
    digits: 1
    specials: 1
    first_is_char: false
  special_chars: "!§$%&/()=?-_+<>|#@;:,.[]{}*"
easy:
  profile:
    length: 8
    upper: 1
    lower: 1
    digits: 1
    specials: 0
strong:
  profile:
    length: 48
    upper: 2
    lower: 2
    digits: 2
    specials: 2
    first_is_char: false
  special_chars: "!§$%&/()=?-_+<>|#@;:,.[]{}*"
````

You may define your own profile sets in a YAML or JSON file with `profile` and
`special_chars` fields.  By default, `password_profiles.yaml` is searched next to the
config file, in the current directory, `$HOME/etc/`, `$HOME/.pwcli`, and `/etc/pwcli`.

Load a custom file with `--password_profiles <filename>`, or set
`password_profiles: <filename>` in the config file.

Example custom profile set:

````yaml
myprofile:
  # Length Upper Lower Digits Specials FirstIsChar
  profile:
    length: 16
    upper: 1
    lower: 1
    digits: 1
    specials: 0
    first_is_char: true
  special_chars: "!#=@&()"
myprofile2:
  ...
````

---

## Command Reference

### Global flags

```
Usage:
  pwcli [command]

Available Commands:
  check       checks a password to given profile
  completion  Generate the autocompletion script for the specified shell
  config      handle config settings
  decrypt     Decrypt crypted file
  encrypt     Encrypt plaintext file
  genkey      Generate a new RSA Keypair
  genpass     generate new password for the given profile
  get         Get encrypted password
  gopass      Manage gopass password store
  hash        commands related to hashing Passwords
  help        Help about any command
  ldap        commands related to ldap
  list        list passwords
  totp        generate totp code from secret
  vault       handle vault functions
  version     version print version string

Global Flags:
  -a, --app string       name of application (default "pwcli")
      --config string    config file name (default "pwcli.yaml")
  -D, --datadir string   directory of password files
      --debug            verbose debug output
      --info             reduced info output
  -K, --keydir string    directory of keys
  -m, --method string    encryption method (openssl|go|enc|plain|vault|kms|age|gpg|gopass) (default "openssl")
      --no-color         disable colored log output
      --no-prompt        disable interactive prompts; return an error instead (for batch use)
```

### config

```
Allows read and write application config

Usage:
  pwcli config [command]

Available Commands:
  get         return value for key of running config
  print       print current config in json format
  save        save current config parameter to file
```

```
pwcli config get — return value for key of running config

Usage:
  pwcli config get [flags] [key]

Flags:
  -h, --help         help for get
  -k, --key string   key to get
```

```
pwcli config print — print current config in json format

Usage:
  pwcli config print [flags]

Aliases:
  print, list, show

Flags:
  -h, --help   help for print
```

```
pwcli config save — save current config parameter to file

Usage:
  pwcli config save [flags]

Flags:
  -f, --filename string   FileName to write
      --force             force overwrite
  -h, --help              help for save
```

### encrypt / decrypt / list

```
pwcli encrypt — Encrypt a plain file

Usage:
  pwcli encrypt [flags]

Flags:
  -c, --crypted string        alternate crypted file
  -h, --help                  help for encrypt
  -p, --keypass string        dedicated password for the private key
      --kms_endpoint string   KMS Endpoint Url
      --kms_keyid string      KMS KeyID
  -t, --plaintext string      alternate plaintext file
```

```
pwcli decrypt — Decrypt a crypted file

Usage:
  pwcli decrypt [flags]

Flags:
  -c, --crypted string        alternate crypted file
  -h, --help                  help for decrypt
  -p, --keypass string        dedicated password for the private key
      --kms_endpoint string   KMS Endpoint Url
      --kms_keyid string      KMS KeyID
  -t, --plaintext string      alternate plaintext file
```

```
pwcli list — List all available password records

Usage:
  pwcli list [flags]

Flags:
  -h, --help                  help for list
  -p, --keypass string        dedicated password for the private key
      --kms_endpoint string   KMS Endpoint Url
      --kms_keyid string      KMS KeyID
```

### get

```
pwcli get — Return a password for an account on a system/database

Usage:
  pwcli get [flags]

Flags:
      --case-sensitive        match user and db/system case sensitive (true for methods vault and gopass)
  -d, --db string             name of the system/database
  -E, --entry string          vault secret entry key within method vault, use together with path
  -h, --help                  help for get
  -p, --keypass string        password for the private key
      --kms_endpoint string   KMS Endpoint Url
      --kms_keyid string      KMS KeyID
  -l, --list                  list all entries like pwcli list
  -P, --path string           vault path to the secret, eg /secret/data/... within method vault, use together with path
  -s, --system string         name of the system/database
  -u, --user string           account/user name
  -A, --vault_addr string     VAULT_ADDR Url (default "$VAULT_ADDR")
  -T, --vault_token string    VAULT_TOKEN (default "$VAULT_TOKEN")
```

### genkey

```
pwcli genkey — Generates a new pair of keys (ecdsa, rsa, age, gpg)
optionally you may assign an individual key password using -p flag
For age keys --keypass creates a passphrase-encrypted private key file (age scrypt).

Usage:
  pwcli genkey [flags]

Flags:
  -h, --help             help for genkey
  -p, --keypass string   dedicated password for the private key
  -t, --type string      key type: ecdsa|rsa|age|gpg (default "rsa")
```

### genpass / checkpass

```
pwcli genpass — generate a random password according the given profile

Usage:
  pwcli genpass [flags]

Aliases:
  genpass, gen, new

Flags:
  -h, --help                       help for genpass
  -l, --list_profiles              list existing profiles only
      --password_profiles string   filename for loading password profile sets
  -p, --profile string             set profile string as numbers of 'Length Upper Lower Digits Special FirstIsCharFlag(0/1)'
  -P, --profileset string          set profile to existing named profile set
  -s, --special_chars string       define allowed special chars
```

```
pwcli checkpass — Checks a password for charset and length rules

Usage:
  pwcli checkpass [flags]

Aliases:
  checkpass, check

Flags:
  -h, --help                       help for check
  -l, --list_profiles              list existing profiles only
      --password_profiles string   filename for loading password profile sets
  -p, --profile string             set profile string as numbers of 'Length Upper Lower Digits Special FirstIsCharFlag(0/1)'
  -P, --profileset string          set profile to existing named profile set
  -s, --special_chars string       define allowed special chars
```

### totp

```
pwcli totp — generate a standard 6-digit auth/MFA code for given secret

Usage:
  pwcli totp [flags]

Flags:
  -h, --help            help for totp
  -s, --secret string   totp secret to generate code from
```

### vault

```
pwcli vault — Allows list, read and write vault secrets

Usage:
  pwcli vault [command]

Available Commands:
  list        list secrets
  read        read a vault secret
  write       write json to vault path

Flags:
  -h, --help                 help for vault
  -L, --logical              Use Logical API (e.g. for database secrets engine), default is KV2
  -M, --mount string         Mount Path of the Secret engine (default "secret/")
  -P, --path string          Vault secret Path to Read/Write
  -A, --vault_addr string    VAULT_ADDR Url (default "$VAULT_ADDR")
  -T, --vault_token string   VAULT_TOKEN (default "$VAULT_TOKEN")
```

```
pwcli vault read — read a secret from given path in KV2 or Logical mode
list all data below path in list_password syntax or give a key as extra arg to return only this value

Usage:
  pwcli vault read [flags]

Flags:
  -E, --export   output as bash export
  -h, --help     help for read
  -J, --json     output as json
```

To retrieve dynamic database credentials from Vault, use the `--logical` flag:

```bash
pwcli vault read --logical --path database/creds/my-role
pwcli vault read --logical --path database/creds/my-role --json
pwcli vault read --logical --path database/creds/my-role --export
# export USERNAME="v-token-my-role-..."
# export PASSWORD="p-..."
```

```
pwcli vault secrets — list secrets recursively below given path (without content)

Usage:
  pwcli vault secrets [flags]

Aliases:
  secrets, list, ls

Flags:
  -h, --help   help for secrets
```

```
pwcli vault write — write a secret to given path in KV2 or Logical mode with json encoded data

Usage:
  pwcli vault write [flags]

Flags:
      --data_file string   Path to the json encoded file with the data to read from
  -h, --help               help for write
```

### gopass

```
pwcli gopass — Manage gopass password store

Usage:
  pwcli gopass [command]

Available Commands:
  identity    Manage age and GPG identity files
  list        List secrets in store
  pull        Git pull the store directory
  push        Git push the store directory
  read        Decrypt and print a secret
  recipients  Manage store recipients
  stores      List configured gopass stores from gopass config
  write       Encrypt and store a secret

Flags:
  -C, --crypto string         Encryption type: age or gpg (auto-detected if empty)
  -h, --help                  help for gopass
      --identity-dir string   Age identity directory for auto-detection (default: ~/.config/gopass/identities/)
  -k, --key-file string       Age identity file (read) or recipients file (write)
  -S, --store-dir string      Path to gopass store directory (auto-detected if empty)
```

```
pwcli gopass read — Decrypt and print a secret

Usage:
  pwcli gopass read <secret> [flags]

Flags:
  -h, --help              help for read
      --keypass string    Passphrase for encrypted age identity file
      --raw               Output full raw secret content instead of first line only
```

```
pwcli gopass write — Encrypt and store a secret

Usage:
  pwcli gopass write <secret> [flags]

Flags:
      --content string   Secret content to store (reads from stdin if not set)
  -h, --help             help for write
```

```
pwcli gopass stores — List configured gopass stores from gopass config

Usage:
  pwcli gopass stores [flags]

Flags:
  -h, --help   help for stores
```

```
pwcli gopass recipients — Manage store recipients

Usage:
  pwcli gopass recipients [command]

Available Commands:
  add         Append a public key to the recipients file
  list        List recipients in store (.age-recipients or .gpg-id)

Flags:
  -h, --help   help for recipients
```

```
pwcli gopass identity — Manage age and GPG identity files

Usage:
  pwcli gopass identity [command]

Available Commands:
  add     Copy an age private key into the identity directory
  create  Generate a new age or GPG key pair and store it in the identity directory
  list    List age identity files in identity directory

Flags:
  -h, --help   help for identity
```

```
pwcli gopass identity create — Generate a new age or GPG key pair and store it in the identity directory

Usage:
  pwcli gopass identity create <alias> [flags]

Flags:
      --add-recipient       Append the new public key to the store recipients file
      --comment string      GPG identity comment
      --email string        GPG identity email
  -h, --help                help for create
      --name string         GPG identity name
      --passphrase string   Private key passphrase (age: scrypt-encrypted key file; GPG: key passphrase)
```

```
pwcli gopass identity add — Copy an age private key into the identity directory

Usage:
  pwcli gopass identity add <alias> <keyfile> [flags]

Flags:
  -h, --help   help for add
```

```
pwcli gopass pull — Git pull the store directory

Usage:
  pwcli gopass pull [flags]

Flags:
  -h, --help            help for pull
      --remote string   Git remote name (default "origin")
```

```
pwcli gopass push — Git push the store directory

Usage:
  pwcli gopass push [flags]

Flags:
  -h, --help            help for push
      --remote string   Git remote name (default "origin")
```

### ldap

```
pwcli ldap — commands related to ldap

Usage:
  pwcli ldap [command]

Available Commands:
  groups      Show the group memberships of the given DN
  members     Search the members of the given group CN
  setpass     change LDAP Password for given User per DN
  setssh      Set public SSH Key to LDAP DN
  show        Show attributes of LDAP DN

Flags:
  -h, --help                       help for ldap
  -b, --ldap.base string           Ldap Base DN
  -B, --ldap.binddn string         DN of user for LDAP bind or use Env LDAP_BIND_DN
  -p, --ldap.bindpassword string   password for LDAP Bind User or use Env LDAP_BIND_PASSWORD
  -H, --ldap.host string           Hostname of Ldap Server
  -I, --ldap.insecure              do not verify TLS
  -P, --ldap.port int              ldap port to connect
  -T, --ldap.targetdn string       DN of target User for admin executed password change, empty for own entry (uses LDAP_BIND_DN)
  -U, --ldap.targetuser string     uid to search for targetDN
      --ldap.timeout int           ldap timeout in sec (default 20)
      --ldap.tls                   use secure ldap (ldaps)
```

```
pwcli ldap setpass — set new ldap password for the actual bind DN or as admin bind for a target DN
if no new password is given some systems will generate a password

Usage:
  pwcli ldap setpass [flags]

Aliases:
  setpass, change-password

Flags:
  -g, --generate                   generate a new password (alternative to being prompted)
  -h, --help                       help for setpass
  -n, --new-password string        new password to set or use Env LDAP_NEW_PASSWORD or be prompted
      --password_profiles string   filename for loading password profile sets
      --profile string             set profile string as numbers of 'Length Upper Lower Digits Special FirstIsCharFlag(0/1)'
      --profileset string          set profile to existing named profile set
```

```
pwcli ldap setssh — set new ssh public key (attribute sshPublicKey) for a given User per DN

Usage:
  pwcli ldap setssh [flags]

Aliases:
  setssh, change-sshpubkey

Flags:
  -h, --help                        help for setssh
  -f, --sshpubkeyfile string        filename with ssh public key to upload (default "id_rsa.pub")
```

```
pwcli ldap show — Show attributes of LDAP DN (own bind DN or searched user)

Usage:
  pwcli ldap show [flags]

Aliases:
  show, show-attributes, attributes

Flags:
  -A, --attributes string   comma separated list of attributes to show (default "*")
  -h, --help                help for show
```

```
pwcli ldap groups — Show the group membership of own User (Bind User) or a searched user

Usage:
  pwcli ldap groups [flags]

Aliases:
  groups, show-groups, group-membership

Flags:
  -h, --help                    help for groups
  -G, --ldap.groupbase string   Base DN for group search
```

### hash

```
pwcli hash — prepare a password hash
currently supports basic auth (for http), md5 and scram (for postgresql),
SSHA (for LDAP), bcrypt (for htpasswd) and argon2 (for vaultwarden)

Usage:
  pwcli hash [command]

Available Commands:
  argon2      command to hashing Passwords with Argon2 method
  basic       command to encoding User/Password with HTTP Basic method
  bcrypt      command to hashing Passwords with BCrypt method
  md5         command to hashing User/Password with MD5 method
  scram       command to hashing User/Password with SCRAM method
  ssha        command to hashing Passwords with SSHA method

Flags:
  -h, --help   help for hash
```

---

## Examples

### Local store

Set up a config, generate a key pair, encrypt a password file and retrieve a password:

```bash
$ pwcli config save -a myapp -D ~/.pwcli -K ~/.pwcli -m go
DONE

$ pwcli genkey -a myapp -m go
DONE

$ cat ~/.pwcli/myapp.plain
prod-db:appuser:s3cr3t
!default:appuser:fallback

$ pwcli encrypt -a myapp
DONE

$ pwcli get -a myapp -u appuser -d prod-db
s3cr3t

$ pwcli get -a myapp -u appuser -d staging
fallback
```

Passphrase-protected key — prompted when `--keypass` is omitted:

```bash
$ pwcli genkey -a myapp -m go --keypass hunter2
DONE

$ pwcli get -a myapp -u appuser -d prod-db
Key passphrase: ••••••••
s3cr3t

$ pwcli get -a myapp -u appuser -d prod-db --keypass hunter2
s3cr3t

$ pwcli --no-prompt get -a myapp -u appuser -d prod-db
Error: key decryption failed: …
```

### gopass store

Bootstrap a store with a fresh age identity, write a secret and read it back:

```bash
$ pwcli gopass identity create mykey --add-recipient -S ~/secrets
identity mykey created [age]
  private key: ~/.config/gopass/identities/mykey.key
  public key:  ~/.config/gopass/identities/mykey.pub
  public key value: age1qyg…
  added to recipients: ~/secrets/.age-recipients

$ pwcli gopass write infra/db/prod -S ~/secrets --content "s3cr3t"
secret infra/db/prod written

$ pwcli gopass read infra/db/prod -S ~/secrets
s3cr3t

$ pwcli gopass list -S ~/secrets
infra/db/prod
```

With a passphrase-protected identity:

```bash
$ pwcli gopass identity create mykey --passphrase hunter2 --add-recipient -S ~/secrets
$ pwcli gopass read infra/db/prod -S ~/secrets
Age identity passphrase: ••••••••
s3cr3t

$ pwcli gopass read infra/db/prod -S ~/secrets --keypass hunter2
s3cr3t
```

Use the gopass store as a backend for `get`:

```bash
$ pwcli get --method gopass -S ~/secrets --path infra/db/prod --entry password
s3cr3t
```

### Vault

```bash
$ export VAULT_ADDR="http://localhost:8200"
$ export VAULT_TOKEN="vault-test"

$ pwcli vault write -P infra/db '{"password":"s3cr3t","user":"appuser"}'
OK

$ pwcli vault read -P infra/db
infra/db:password:s3cr3t
infra/db:user:appuser

$ pwcli vault read -P infra/db --json
{"password":"s3cr3t","user":"appuser"}

$ pwcli vault read -P infra/db password
s3cr3t

$ pwcli vault secrets -P /
infra/db

# Dynamic database credentials via logical API
$ pwcli vault read --logical --path database/creds/my-role --json
{"lease_duration":3600,"username":"v-token-my-role-abc","password":"p-xyz"}

$ pwcli vault read --logical --path database/creds/my-role --export
export USERNAME="v-token-my-role-abc"
export PASSWORD="p-xyz"

# get command via vault method
$ pwcli get --method vault --path infra/db --entry password
s3cr3t
```

### KMS

```bash
$ export AWS_REGION="eu-central-1"
$ export AWS_ACCESS_KEY_ID="..."
$ export AWS_SECRET_ACCESS_KEY="..."
$ export KMS_KEYID="alias/mykey"

$ pwcli encrypt -a myapp -m kms
DONE

$ pwcli get -a myapp -m kms -u appuser -d prod-db
s3cr3t
```

### Password profiles

```bash
# Generate with inline profile: length=16, 1 upper, 1 lower, 1 digit, 1 special, first is letter
$ pwcli genpass --profile "16 1 1 1 1 1" --special_chars '#!@=?'
R3qu!reXmNp=4Lk7

# Use a named profileset
$ pwcli genpass --profileset devk_user
Qh7#mNpL

# List available profilesets
$ pwcli genpass -l
default: length=16 …
devk_user: length=12 …
easy: length=8 …
strong: length=48 …

# Validate a password against a profileset
$ pwcli checkpass --profileset devk_user "Qh7#mNpL"
SUCCESS

$ pwcli checkpass --profileset devk_user "abc"
ERROR length check failed: at least 12 chars expected, have 3
password 'abc' matches NOT the given profile
```

### LDAP

```bash
# Set password interactively
$ pwcli ldap setpass -H ldap.example.com -P 389 \
    -B cn=alice,ou=Users,dc=example,dc=com -p currentpass
Change password for cn=alice,ou=Users,dc=example,dc=com
Enter NEW password: ••••••••
Repeat NEW password: ••••••••
Password for cn=alice,ou=Users,dc=example,dc=com changed and tested

# Set password non-interactively
$ pwcli ldap setpass -H ldap.example.com -P 389 \
    -B cn=alice,ou=Users,dc=example,dc=com -p currentpass -n newpass
Password for cn=alice,ou=Users,dc=example,dc=com changed and tested

# Generate a new password and apply it
$ pwcli ldap setpass -H ldap.example.com -P 389 \
    -B cn=alice,ou=Users,dc=example,dc=com -p currentpass -g --profileset devk_user
generated Password: Qh7#mNpL9xRt
Password for cn=alice,ou=Users,dc=example,dc=com changed and tested

# Upload an SSH public key
$ pwcli ldap setssh -H ldap.example.com -P 389 \
    -B cn=alice,ou=Users,dc=example,dc=com -p currentpass -f ~/.ssh/id_ed25519.pub
SSH Key for cn=alice,ou=Users,dc=example,dc=com changed

# Show selected attributes
$ pwcli ldap show -H ldap.example.com -P 389 \
    -B cn=alice,ou=Users,dc=example,dc=com -p currentpass -A cn,mail,sshPublicKey
DN 'cn=alice,ou=Users,dc=example,dc=com' has following attributes:
cn: alice
mail: alice@example.com
sshPublicKey: ssh-ed25519 AAAA…

# Show group membership
$ pwcli ldap groups -H ldap.example.com -P 389 \
    -B cn=alice,ou=Users,dc=example,dc=com -p currentpass
DN 'cn=alice,ou=Users,dc=example,dc=com' is member of the following groups:
Group: cn=developers,ou=Groups,dc=example,dc=com
Group: cn=ssh,ou=Groups,dc=example,dc=com

# List members of a group
$ pwcli ldap members -H ldap.example.com -P 389 \
    -B cn=alice,ou=Users,dc=example,dc=com -p currentpass -g developers
search for 'developers' members in following groups:
Group: cn=developers,ou=Groups,dc=example,dc=com
    Member: cn=alice,ou=Users,dc=example,dc=com
    Member: cn=bob,ou=Users,dc=example,dc=com
```

### Hash

```bash
# BCrypt (htpasswd)
$ pwcli hash bcrypt --password=mypass
$2a$10$…

$ pwcli hash bcrypt --password=mypass --test='$2a$10$…'
OK, test input matches bcrypt hash

# SSHA (LDAP)
$ pwcli hash ssha --password=mypass
{SSHA}o0jvU/LY4KFsq5MgUtc0aB/KQY3QfrFH

# SCRAM-SHA-256 (PostgreSQL)
$ pwcli hash scram --username=appuser --password=mypass
SCRAM-SHA-256$4096:…

# MD5 (legacy)
$ pwcli hash md5 --username=appuser --password=mypass
md5ed2dbc3fbef8ab0b846185e442fd0ce2

# Argon2id (vaultwarden / general)
$ pwcli hash argon2 -p mypass
$argon2id$v=19$m=65536,t=3,p=4$…

# HTTP Basic Auth header
$ pwcli hash basic --username=appuser --password=mypass
Authorization: Basic YXBwdXNlcjpteXBhc3M=
```

### TOTP

```bash
$ pwcli totp --secret "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
134329

# Avoid leaking the secret to shell history via environment variable
$ export TOTP_SECRET="GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
$ pwcli totp
197004
```
