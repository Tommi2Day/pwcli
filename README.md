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
  version     version print version string

Flags:
  -a, --app string       name of application (default "pwcli")
      --config string    config file name (default "pwcli.yaml")
  -D, --datadir string   directory of password files
      --debug            verbose debug output
  -h, --help             help for pwcli
      --info             reduced info output
  -K, --keydir string    directory of keys
  -m, --method string    encryption method (openssl|go|enc|plain) (default "go")

Use "pwcli [command] --help" for more information about a command.

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

> pwcli config save -a test_pwcli -D test/testdata -K test/testdata -m go
DONE

> pwcli pwcli genkey -a test_pwcli --keypass pwcli_test --info
[Thu, 20 Apr 2023 14:33:38 CEST]  INFO New key pair generated as test/testdata/test_pwcli.pub and test/testdata/test_pwcli.pem
DONE

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

> pwcli list -a test_pwcli -p pwcli_test
> pwcli list -a test_pwcli -p pwcli_test --info
# Testfile
!default:defuser2:failure
!default:testuser:default
test:testuser:testpass
testdp:testuser:xxx:yyy
!default:defuser2:default
!default:testuser:failure
!default:defuser:default

[Thu, 20 Apr 2023 14:37:38 CEST]  INFO List returned 10 lines

> pwcli get -a test_pwcli -u testuser -d test -m plain -D test/testdata
testpass
> pwcli get -a test_pwcli -u testuser -d test -D test/testdata -K test/testdata -m go -p pwcli_test
testpass

> pwcli new --profile "16 1 1 1 1 1" --special_chars '#!@=?'
iFTQVxT==CV#6X1k

> pwcli check -p "8 1 1 1 0 1" "1234abc"
[Thu, 20 Apr 2023 14:38:47 CEST] ERROR length check failed: at least  8 chars expected, have 7
[Thu, 20 Apr 2023 14:38:47 CEST] ERROR uppercase check failed: at least 1 chars out of 'ABCDEFGHIJKLMNOPQRSTUVWXYZ' expected
[Thu, 20 Apr 2023 14:38:47 CEST] ERROR first character check failed: first letter check failed, only ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijkmlnopqrstuvwxyz allowed
password '1234abc' matches NOT the given profile

```


