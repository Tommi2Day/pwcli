# pwcli

Toolbox for validating, storing and query encrypted passwords

[![Go Report Card](https://goreportcard.com/badge/github.com/tommi2day/pwcli)](https://goreportcard.com/report/github.com/tommi2day/pwcli)
![CI](https://github.com/tommi2day/pwcli/actions/workflows/main.yml/badge.svg)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tommi2day/pwcli)

## Usage

```bash
Usage:
  pwcli [command]

Available Commands:
  check       checks a password to given profile
  completion  Generate the autocompletion script for the specified shell
  config      handle config settings
  encrypt     Encrypt plaintext file
  genkey      Generate a new RSA Keypair
  genpass     generate new password for the given profile
  get         Get encrypted password
  help        Help about any command
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
  -m, --method string    encryption method (openssl|go|enc|plain|vault) (default "openssl")
  
Use "pwcli [command] --help" for more information about a command.
#-------------------------------------
pwcli config save --help
write application config

Usage:
  pwcli config save [flags]

Flags:
  -f, --filename string   FileName to write
      --force             force overwrite
  -h, --help              help for save
#-------------------------------------
pwcli check --help
Checks a password for charset and length rules

Usage:
  pwcli check [flags]

Flags:
  -h, --help                   help for check
  -p, --profile string         set profile string as numbers of 'length Upper Lower Digits Special FirstcharFlag(0/1)' (default "10 1 1 1 0 1")
  -s, --special_chars string   define allowed special chars (default "!?()-_=")
#-------------------------------------
pwcli genkey --help
Generates a new pair of rsa keys
optionally you may assign an idividal key password using -p flag

Usage:
  pwcli genkey [flags]

Aliases:
  genkey, genrsa

Flags:
  -h, --help             help for genkey
  -p, --keypass string   dedicated password for the private key
#-------------------------------------
pwcli encrypt --help
Encrypt a plain file given in -p and saved as crypted fin given by -c flag
default for paintext File is <app>.plain and for crypted file is <app.pw>

Usage:
  pwcli encrypt [flags]

Flags:
  -c, --crypted string     alternate crypted file
  -h, --help               help for encrypt
  -p, --keypass string     dedicated password for the private key
  -t, --plaintext string   alternate plaintext file

#-------------------------------------
pwcli genpass --help
this will generate a random password according the given profile

Usage:
  pwcli genpass [flags]

Aliases:
  genpass, gen, new

Flags:
  -h, --help                   help for genpass
  -p, --profile string         set profile string as numbers of 'length Upper Lower Digits Special FirstcharFlag(0/1)' (default "10 1 1 1 0 1")
  -s, --special_chars string   define allowed special chars (default "!?()-_=")
#-------------------------------------
pwcli get --help
Return a password for a an Account on a system/database
Usage:
  pwcli get [flags]

Flags:
  -d, --db string            name of the system/database
  -E, --entry string         vault secret entry key within method vault, use together with path
  -h, --help                 help for get
  -p, --keypass string       password for the private key
  -P, --path string          vault path to the secret, eg /secret/data/... within method vault, use together with path
  -s, --system string        name of the system/database
  -u, --user string          account/user name
  -A, --vault_addr string    VAULT_ADDR Url (default "$VAULT_ADDR")
  -T, --vault_token string   VAULT_TOKEN (default "$VAULT_TOKEN")

#-------------------------------------
pwcli totp --help
generate a standard 6 digit auth/mfa code for given secret with --secret or TOTP_SECRET env

Usage:
  pwcli totp [flags]

Flags:
  -h, --help            help for totp
  -s, --secret string   totp secret to generate code from
#-------------------------------------
pwcli vault --help 
Allows list, read and write vault secrets

Usage:
  pwcli vault [command]

Available Commands:
  list        list secrets
  read        read a vault secret
  write       write json to vault path

Flags:
  -h, --help                 help for vault
  -L, --logical              Use Logical Api, default is KV2
  -M, --mount string         Mount Path of the Secret engine (default "secret/")
  -P, --path string          Vault secret Path to Read/Write
  -A, --vault_addr string    VAULT_ADDR Url (default "$VAULT_ADDR")
  -T, --vault_token string   VAULT_TOKEN (default "$VAULT_TOKEN")

```
### format plaintext file
plaintextfile should be named as `<app>.plain` and stored in `datadir` to encrypt
```
# system:user:password
!default:testuser:default # default match for this user on each system
test:testuser:testpass    # exact match, has precedence over default
```

## Examples
```bash
> pwcli version

# save the config parameters to file, will be a preset for appname test_pwcli
> pwcli config save -a test_pwcli -D test/testdata -K test/testdata -m go
DONE

# generate a new keypair for test_pwcli and set passphrase via --keypass 
> pwcli genkey -a test_pwcli --keypass pwcli_test --info
[Thu, 20 Apr 2023 14:33:38 CEST]  INFO New key pair generated as test/testdata/test_pwcli.pub and test/testdata/test_pwcli.pem
DONE

# encrypt a file named test_pwcli.plain with given keyset test_pwcli and keypass
> pwcli encrypt -a test_pwcli --keypass pwcli_test --debug
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG NewConfig entered
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG A:test_pwcli, P:, D:test/testdata, K:test/testdata, M:go
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG encrypt called
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG encrypt plaintext file 'test/testdata/test_pwcli.plain' with method go
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG create crypted file 'test/testdata/test_pwcli.gp'
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG use alternate key password ''
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG Encrypt data from test/testdata/test_pwcli.plain method go
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG Encrypt test/testdata/test_pwcli.plain with public key test/testdata/test_pwcli.pub
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG load public key from test/testdata/test_pwcli.pub
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG public key loaded successfully
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG Session key len: 256
[Thu, 20 Apr 2023 14:36:07 CEST] DEBUG encrytion data success
[Thu, 20 Apr 2023 14:36:07 CEST]  INFO crypted file 'test/testdata/test_pwcli.gp' successfully created
DONE

# list encrypted passwords and handover RSA key passphrase
> pwcli list -a test_pwcli -p pwcli_test
pwcli list -a test_pwcli -p pwcli_test --info
# Testfile
!default:defuser2:failure
!default:testuser:default
test:testuser:testpass
testdp:testuser:xxx:yyy
!default:defuser2:default
!default:testuser:failure
!default:defuser:default

[Thu, 20 Apr 2023 14:37:38 CEST]  INFO List returned 10 lines

# retrieve a password via diffenent methods
> pwcli get -a test_pwcli -u testuser -d test -m plain -D test/testdata
testpass
> pwcli get -a test_pwcli -u testuser -d test -D test/testdata -K test/testdata -m go -p pwcli_test
testpass

# create vault demo entry
>vault kv put secret/mysecret password=testpass
# query vault this secret using logical path
>pwcli get --method=vault --path=secret/data/mysecret --entry=password

# generate a random password with 16 chars, 
#   at least one upper,
#   one lower char, 
#   one digit 
#   one special char from give charset
#   and use no digits and specials as first char  
> pwcli new --profile "16 1 1 1 1 1" --special_chars '#!@=?'
iFTQVxT==CV#6X1k

# check value not matching the given profile
> pwcli check -p "8 1 1 1 0 1" "1234abc"
[Thu, 20 Apr 2023 14:38:47 CEST] ERROR length check failed: at least  8 chars expected, have 7
[Thu, 20 Apr 2023 14:38:47 CEST] ERROR uppercase check failed: at least 1 chars out of 'ABCDEFGHIJKLMNOPQRSTUVWXYZ' expected
[Thu, 20 Apr 2023 14:38:47 CEST] ERROR first character check failed: first letter check failed, only ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijkmlnopqrstuvwxyz allowed
password '1234abc' matches NOT the given profile

# write vault secret direct to KV and handover VAULT_ADDR and VAULT_TOKEN via cli
>pwcli vault write -P test '{"password": "secret"}' -A "http://localhost:8200" -T "vault-test"
OK

# list in pwlib plain format
>pwcli vault read -P test -A "http://localhost:8200" -T "vault-test"
test:password:secret

# list as json
>pwcli vault read -P test -A "http://localhost:8200" -T "vault-test" -J
{"password":"secret"}

# read single key with Vault environment set, 
# see also pwcli get command in vault mode
export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN="vault-test"
pwcli vault read -P test "password"
secret

# list from valid path
pwcli vault list -P "/" --info
[Wed, 26 Apr 2023 17:37:25 CEST]  INFO Vault List returned 1 entries
test

# list from invalid path
>pwcli vault list -P "xx"
no Entries returned

# generate totp value from given secret
>pwcli totp --secret "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
134329

# use environment to provide the secret, avoid commandline
>export TOTP_SECRET="GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
>pwcli totp
197004

# wrong secret
>pwcli totp --secret "xxx"
TOTP generation failed:panic:decode secret failed
```


