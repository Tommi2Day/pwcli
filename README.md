# pwcli

Toolbox for validating, storing and query encrypted passwords

[![Go Report Card](https://goreportcard.com/badge/github.com/tommi2day/pwcli)](https://goreportcard.com/report/github.com/tommi2day/pwcli)
![CI](https://github.com/tommi2day/pwcli/actions/workflows/main.yml/badge.svg)
[![codecov](https://codecov.io/gh/Tommi2Day/pwcli/branch/main/graph/badge.svg?token=3EBK75VLC8)](https://codecov.io/gh/Tommi2Day/pwcli)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tommi2day/pwcli)

## Features
this tool contains a collection of often used solution for
- using encrypted files or password stores for password retrieving with different methods
  - with rsa keys (default)
    - with go crypto functions
    - with openssl compatible format (default)
  - with kms keys
  - from hashicorp vault
  - from (unsecure) plain files
  - from base64 encoded files

- generate and check secure passwords with password profiles

- generate hashes with common used methods and test if a given user matches the hash (if possible)
  - md5
  - ssha (e.g. for LDAP Passwords)
  - argon2
  - bcrypt
  - http basic auth format
  - SCRAM (e.g. for postgresql)

- generate RSA and KMS Keys

- generate TOTP codes

- ldap functions
  - set ldap password
  - set sshkey field
  - retrieve entries
  - lookup group membership of user
  - lookup users by group

## Use pwcli as password store
you may have different password stores using different formats, keys and directories.
the identifier ist the `application` name

## config files
this tool may use config files via viper. the default filename is `<application>.yaml` and fallback to `pwcli.yaml`
It will search
- in the current directory
- in $HOME/.pwcli,
- in $HOME/etc/
- /etc/pwcli
  You may define your own config file path with --config switch

this generates a default configfile `get_password.yaml` for the application `get_password` in the current directory
and configures the datadir and keydir to `$HOME/.pwcli` and the encryption method RSA in GO file format
````shell
pwcli config save -a get_password -D $HOME/.pwcli -K $HOME/.pwcli -m go
````
you may mention this config filename in all commands
make sure the directories in -D and -K exists
````shell
mkdir $HOME/.pwcli
````
### generate rsa keys used for encryption
as we will use RSA encryption we have to create a corresponding keypair.
you may set your own keypass using --keypass option, or it's using its own
the following uses the configuration file above to generate the needed keys within the correct directories
````shell
pwcli genkey -a get_password
````
### password store file
if not using a third party passwordstore as is hashicorp vault or gopass the local password store is derived from
plaintext files and encrypted via the method as given in config or switch

these files following a special format

`system:user:password`

a special system name `!default` indicates it returns the password for all systems to given user

the plaintext file should be named as `<app>.plain` and stored in `datadir` to encrypt
or if not following this convention refer this file with --plaintext option
```
# default match for this user on each system
!default:testuser:default
# exact match, has precedence over default
test:testuser:testpass
```

````shell
pwcli encrypt -a get_password --plaintext test.plain
````
### Query password
with the encrypted file you may query the desired password.

search for system and user is case-insensitive per default, except for methods vault and gopasse,
but you may activate case-sensitivity for all methods using `--case-sensitive` switch
````shell
pwcli get -a get_password -u testuser -d test
testpass
````
## password profile sets
there are some profileset predefined. These can be used to generate or check a password with a named rule. If no passwordset or former profile string is given
the default profileset will be taken
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
you may define your own password profile sets. It should be a similar yaml or json file with profile and special_chars setting.
Per default, it will look for a password_profiles.yaml
- next to the configfile,
- in the current directory,
- in $HOME/etc/
- in $HOME/.pwcli
- /etc/pwcli

You may also write your own file elsewhere. This file can be loaded via option `--password_profiles <filename>` or
automatic with `password_profiles: <filename>` entry in configfile
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

## Command Reference

```bash
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
  -m, --method string    encryption method (openssl|go|enc|plain|vault|kms) (default "openssl")
      --no-color         disable colored log output
      --unit-test        redirect output for unit tests

Use "pwcli [command] --help" for more information about a command.
#-------------------------------------
pwcli config --help
Allows read and write application config


Usage:
  pwcli config [command]

Available Commands:
  get         return value for key of running config
  print       print current config in json format
  save        save current config parameter to file

#-----------------------------------  
pwcli config get --help
return value for key of running config, pass the viper key as argument or using -k flag

Usage:
  pwcli config get [flags] [key]

Flags:
  -h, --help         help for get
  -k, --key string   key to get
#-----------------------------------
pwcli config print --help
print current config in json format

Usage:
  pwcli config print [flags]

Aliases:
  print, list, show

Flags:
  -h, --help   help for print

#-----------------------------------
pwcli config save --help
save current config parameter to file

Usage:
  pwcli config save [flags]

Flags:
  -f, --filename string   FileName to write
      --force             force overwrite
  -h, --help              help for save
#-------------------------------------
pwcli checkpass --help
Checks a password for charset and length rules

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
#-------------------------------------
pwcli genpass --help
this will generate a random password according the given profile

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
pwcli decrypt --help
Decrypt a crypted file given in -c and saved as plaintext file given by -p flag using given method.
default for plaintext File is <app>.plain and for crypted file is <app.pw>

Usage:
  pwcli decrypt [flags]

Flags:
  -c, --crypted string        alternate crypted file
  -h, --help                  help for decrypt
  -p, --keypass string        dedicated password for the private key
      --kms_endpoint string   KMS Endpoint Url
      --kms_keyid string      KMS KeyID
  -t, --plaintext string      alternate plaintext file

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
#--------------------------------------
pwcli vault read --help

read a secret from given path in KV2 or Logical mode
list all data below path in list_password syntax or give a key as extra arg to return only this value

Usage:
  pwcli vault read [flags]

Flags:
  -E, --export   output as bash export
  -h, --help     help for read
  -J, --json     output as json
#-------------------------------------
pwcli vault secrets --help
list secrets recursive below given path (without content)

Usage:
  pwcli vault secrets [flags]

Aliases:
  secrets, list, ls

Flags:
  -h, --help   help for secrets
#-------------------------------------
pwcli vault write --help
write a secret to given path in KV2 or Logical mode with json encoded data

Usage:
  pwcli vault write [flags]

Flags:
      --data_file string   Path to the json encoded file with the data to read from
  -h, --help               help for write
#-----------------------------
pwcli ldap --help
commands related to ldap

Usage:
  pwcli ldap [command]

Available Commands:
  groups      Show the group memberships of the given DN
  members     Search the members  of the given group CN
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
  -g, --generate                   generate a new password (alternative to be prompted)
  -h, --help                       help for setpass
  -n, --new-password string        new_password to set or use Env LDAP_NEW_PASSWORD or be prompted
      --password_profiles string   filename for loading password profile sets
      --profile string             set profile string as numbers of 'Length Upper Lower Digits Special FirstIsCharFlag(0/1)'
      --profileset string          set profile to existing named profile set

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
pwcli hash
prepare a password hash
currently supports basic auth(for http), md5 and scram(for postgresql),SSHA(for LDAP), bcrypt(for htpasswd) and argon2(for vaultwarden)

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

# decrypt a file named test_pwcli.gp with given keyset test_pwcli and keypass
>pwcli decrypt -a test_pwcli --debug -t test/testdata/plain.txt -c test/testdata/test_pwcli.gp --method go -K test/testdata/
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG found configfile /home/hv11647/etc/pwcli.yaml
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG NewConfig entered
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG decrypt called
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG decrypt file 'test/testdata/test_pwcli.gp' with method go
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG create plaintext file 'test/testdata/plain.txt'
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG use alternate key password ''
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG Decrypt data from test/testdata/test_pwcli.gp with method go(Encypted)
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG decrypt test/testdata/test_pwcli.gp with private key test/testdata//test_pwcli.pem
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG file test/testdata/test_pwcli.gp exists

[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG GetPrivateKeyFromFile entered for test/testdata//test_pwcli.pem
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG file test/testdata/test_pwcli.pem exists

[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG Keys successfully loaded
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG Session key decrypted
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG Decoding successfully
[Tue, 24 Sep 2024 18:54:53 CEST] DEBUG load data success
[Tue, 24 Sep 2024 18:54:53 CEST]  INFO plaintext file 'test/testdata/plain.txt' successfully created
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
# list available profilesets
>pwcli genpass -l
...
easy:
    profile:
        length: 8
        upper: 1
        lower: 1
        digits: 1
        first_is_char: false
# use this  profilesets
>pwcli genpass --profileset easy
8iCUzW2U
# use your own profileset
>pwcli genpass --password_profiles my_profilesets.yaml --profileset myprofile
E!TNmNQEc2Phl5!d

# check value not matching the given profile
> pwcli check -p "8 1 1 1 0 1" "1234abc"
[Thu, 20 Apr 2023 14:38:47 CEST] ERROR length check failed: at least  8 chars expected, have 7
[Thu, 20 Apr 2023 14:38:47 CEST] ERROR uppercase check failed: at least 1 chars out of 'ABCDEFGHIJKLMNOPQRSTUVWXYZ' expected
[Thu, 20 Apr 2023 14:38:47 CEST] ERROR first character check failed: first letter check failed, only ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijkmlnopqrstuvwxyz allowed
password '1234abc' matches NOT the given profile

# check with profilecheck
>pwcli check --profileset easy "8iCUzW2U"
SUCCESS
>pwcli check --password_profiles password_profilesets.yaml.sample --profileset myprofile "8iCUzW2U"
[Sun, 23 Feb 2025 17:36:35 CET] ERROR length check failed: at least  16 chars expected, have 8
[Sun, 23 Feb 2025 17:36:35 CET] ERROR first character check failed: first letter check failed, only ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijkmlnopqrstuvwxyz allowed
password '8iCUzW2U' matches NOT the given profile

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
>pwcli ldap groups -U 'test*'
v cn=test2,ou=Users,dc=example,dc=local
DN 'cn=test2,ou=Users,dc=example,dc=local' is member of the following groups:
Group: cn=ssh,ou=Groups,dc=example,dc=local

# show ldap members of given group
>pwcli ldap members -H localhost -P 2389 --ldap.binddn=cn=test,ou=Users,dc=example,dc=local --ldap.bindpassword test2 -g ssh
search for 'ssh' members in following groups:
Group: cn=ssh,ou=Groups,dc=example,dc=local
    Member: cn=test,ou=Users,dc=example,dc=local
    Member: cn=test2,ou=Users,dc=example,dc=local

Group: cn=ssh-2,ou=Groups,dc=example,dc=local
    Member: cn=test2,ou=Users,dc=example,dc=local

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

>pwcli hash md5 --username=test --password=testpassword --prefix=md5
md5ed2dbc3fbef8ab0b846185e442fd0ce2

>pwcli hash scram --username=test --password=testpassword
SCRAM-SHA-256$4096:SkaOJrvU3G6w2fS2ISDHBGNlCDc99wSS$EZ8KldB0AubHZJOuQL/3HwcxhTYm8P8KqsiG0YsuqRE=:f2uDXFCVABZKO5bP0ZaIQxa247OhGKQ/b/KIwZxfXTQ=

>pwcli hash bcrypt --password=testpassword
$2a$10$e/3qiMq0JrZfsHDS6OnhrORmalQZ7iwkVJWf0HcfxVYWKocFJqObm

>pwcli hash --hash-method ssha --password=testpassword
{SSHA}o0jvU/LY4KFsq5MgUtc0aB/KQY3QfrFH

pwcli hash basic --username=test --password=testpassword
Authorization: Basic dGVzdDp0ZXN0cGFzc3dvcmQ=

>pwcli hash bcrypt --password=testpassword --test='$2a$10$e/3qiMq0JrZfsHDS6OnhrORmalQZ7iwkVJWf0HcfxVYWKocFJqObm'
OK, test input matches bcrypt hash

>pwcli hash argon2 -p testPassword
$argon2id$v=19$m=65536,t=3,p=4$ERt8yTPGDHEV0AZn2yEY/Q$tV9is7EY6KljUuw2vcmDuI0I/BOmIyDoWEj52XU75jI
```
## Virus Warnings

some engines are reporting a virus in the binaries. This is a false positive. You may check the binaries with meta engines such as [virustotal.com](https://www.virustotal.com/gui/home/upload) or build your own binary from source.
I have no glue why this happens.



