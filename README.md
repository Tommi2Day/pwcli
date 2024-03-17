# pwcli

Toolbox for validating, storing and query encrypted passwords

[![Go Report Card](https://goreportcard.com/badge/github.com/tommi2day/pwcli)](https://goreportcard.com/report/github.com/tommi2day/pwcli)
![CI](https://github.com/tommi2day/pwcli/actions/workflows/main.yml/badge.svg)
[![codecov](https://codecov.io/gh/Tommi2Day/pwcli/branch/main/graph/badge.svg?token=3EBK75VLC8)](https://codecov.io/gh/Tommi2Day/pwcli)
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
  hash        commands related to hashing Passwords for use in Postgresql
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
  -m, --method string    encryption method (openssl|go|enc|plain|vault) (default "openssl")
      --no-color         disable colored log output
      --unit-test        redirect output for unit tests

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
  -c, --crypted string        alternate crypted file
  -h, --help                  help for encrypt
  -p, --keypass string        dedicated password for the private key
      --kms_endpoint string   KMS Endpoint Url
      --kms_keyid string      KMS KeyID
  -t, --plaintext string      alternate plaintext file

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
pwcli list --help
List all available password records

Usage:
  pwcli list [flags]

Flags:
  -h, --help                  help for list
  -p, --keypass string        dedicated password for the private key
      --kms_endpoint string   KMS Endpoint Url
      --kms_keyid string      KMS KeyID
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
#-------------------------------------
pwcli ldap --help
commands related to ldap

Usage:
  pwcli ldap [command]

Available Commands:
  groups      Show the group memberships of the given DN
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

#-------------------------------------
pwcli ldap setssh --help
set new ssh public key(attribute sshPublicKey) for a given User per DN, the key must be in a file given by --sshpubkeyfile or default id_rsa.pub.

Usage:
  pwcli ldap setssh [flags]

Aliases:
  setssh, change-sshpubkey

Flags:
  -h, --help                        help for setssh
  -f, --sshpubkeyfile string        filename with ssh public key to upload (default "id_rsa.pub")
#-------------------------------------
pwcli ldap setpass --help
set new ldap password by --new-password or Env LDAP_NEW_PASSWORD for the actual bind DN or as admin bind for a target DN.
if no new password given some systems will generate a password

Usage:
  pwcli ldap setpass [flags]

Aliases:
  setpass, change-password

Flags:
  -h, --help                  help for setpass
  -n, --new-password string   new_password to set or use Env LDAP_NEW_PASSWORD
#-------------------------------------
pwcli ldap show --help
This command shows the attributes off the own User(Bind User) or
you may lookup a User cn and show the attributes of the first entry returned.


Usage:
  pwcli ldap show [flags]

Aliases:
  show, show-attributes, attributes

Flags:
  -A, --attributes string   comma separated list of attributes to show (default "*")
  -h, --help                help for show

#-------------------------------------
pwcli ldap groups --help
This command shows the group membership of  own User(Bind User) or
you may lookup a User cn and if found show the groups of the first entry returned

Usage:
  pwcli ldap groups [flags]

Aliases:
  groups, show-groups, group-membership

Flags:
  -h, --help                    help for groups
  -G, --ldap.groupbase string   Base DN for group search
#-------------------------------------
pwcli hash --help
prepare a password hash
currently supports md5 and scram(for postgresql) and bcrypt(for htpasswd) methods

Usage:
  pwcli hash [flags]

Flags:
      --hash-method string   method to use for hashing, one of md5, scram, bcrypt
  -h, --help                 help for hash
      --password string      password to encrypt
      --username string      username for scram encrypt



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
> pwcli genkey -a test_pwcli --keypass pwcli_test -m go --info
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

# use amazon kms keys (test with local-kms, see test/docker/kms/init/seed.yaml for keyid)
export AWS_REGION="eu-central-1"
export AWS_ACCESS_KEY_ID="test"
export AWS_SECRET_ACCESS_KEY="test"
> pwcli encrypt -a test_pwcli -D test/testdata -K test/testdata -m kms --kms_endpoint "http://localhost:18080" --kms_keyid "bc436485-5092-42b8-92a3-0aa8b93536dc" --info
[Sun, 17 Mar 2024 16:27:46 CET]  INFO crypted file 'test/testdata/test_pwcli.kms' successfully created
DONE
# specify ENDPOINT and KEYID via environment
export KMS_KEYID="alias/testing" 
> pwcli list -a test_pwcli -D test/testdata -K test/testdata -m kms
...
> pwcli get -a test_pwcli -D test/testdata -K test/testdata -m kms -u testuser -s test
testpass

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

# change ldap password for given user and password parameter
>pwcli ldap setpass -H localhost -P 2389 -B=cn=test,ou=Users,dc=example,dc=local -p test -n new_pass
Password for cn=test,ou=Users,dc=example,dc=local changed and tested

# change password and use generated password, default profile 8 1 1 1 0 0
>pwcli ldap setpass -H localhost -P 2389 -B=cn=test,ou=Users,dc=example,dc=local -p test -g --profile "10 1 1 1 1 1"
generated Password: XK8v_hZrdc
Password for cn=test2,ou=Users,dc=example,dc=local changed and tested

# change password without providing password on commandline and enter new password interactively (or supply in env LDDAP_NEW_PASSWORD
>pwcli ldap setpass -H localhost -P 2389 -B=cn=test,ou=Users,dc=example,dc=local -p test
Change password for cn=test2,ou=Users,dc=example,dc=local
Enter NEW password: *****
Repeat NEW password: *****
Password for cn=test,ou=Users,dc=example,dc=local changed and tested

# set new ssh public key for given user
>pwcli ldap setssh -H localhost -P 2389 --ldap.binddn=cn=test,ou=Users,dc=example,dc=local --ldap.bindpassword test2 -f ~/.ssh/id_rsa.pub
SSH Key for cn=test,ou=Users,dc=example,dc=local changed

# show ldap entry attributes
>pwcli ldap show -H localhost -P 2389 --ldap.binddn=cn=test,ou=Users,dc=example,dc=local --ldap.bindpassword test2  -A objectclass,cn,sn
DN 'cn=test,ou=Users,dc=example,dc=local' has following attributes:
cn: test
sn: test
objectClass: top
objectClass: person
objectClass: organizationalPerson
objectClass: inetOrgPerson
objectClass: ldapPublicKey

# show ldap entry group membership and search for user
>pwcli ldap groups -H localhost -P 2389 --ldap.binddn=cn=test,ou=Users,dc=example,dc=local --ldap.bindpassword test -U 'test*'
>pwcli.exe ldap groups -U 'test*'
v cn=test2,ou=Users,dc=example,dc=local
DN 'cn=test2,ou=Users,dc=example,dc=local' is member of the following groups:
Group: cn=ssh,ou=Groups,dc=example,dc=local


# use hashicorp vault 
export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN="vault-test"
# create vault demo entry
>vault kv put secret/mysecret password=testpass
# query vault this secret using logical path
>pwcli get --method=vault --path=secret/data/mysecret --entry=password
# write vault secret direct to KV
>pwcli vault write -P test '{"password": "secret"}' 
OK

# list in pwlib plain format
>pwcli vault read -P test 
test:password:secret

# list as json
>pwcli vault read -P test -J
{"password":"secret"}

# read single key with Vault environment set on commandline
# see also pwcli get command in vault mode
pwcli vault read -P test "password" --vault_addr "http://localhost:8200" --vault_token "vault-test" -J
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

>pwcli.exe hash --hash-method md5 --username=test --password=testpassword
md5ed2dbc3fbef8ab0b846185e442fd0ce2

>pwcli.exe hash --hash-method scram --username=test --password=testpassword
SCRAM-SHA-256$4096:SkaOJrvU3G6w2fS2ISDHBGNlCDc99wSS$EZ8KldB0AubHZJOuQL/3HwcxhTYm8P8KqsiG0YsuqRE=:f2uDXFCVABZKO5bP0ZaIQxa247OhGKQ/b/KIwZxfXTQ=

>pwcli.exe hash --hash-method bcrypt --password=testpassword
$2a$10$e/3qiMq0JrZfsHDS6OnhrORmalQZ7iwkVJWf0HcfxVYWKocFJqObm
```
## Virus Warnings

some engines are reporting a virus in the binaries. This is a false positive. You may check the binaries with meta engines such as [virustotal.com](https://www.virustotal.com/gui/home/upload) or build your own binary from source.
I have no glue why this happens.



