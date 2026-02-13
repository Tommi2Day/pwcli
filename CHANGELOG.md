# Changelog pwcli

# [v2.18.0 - 2026-02-14]
### New
- add vault read --export flag to export secrets in bash environment variable format
- add more vault tests
### Changed
- update dependencies
- rename vault list to vault secrets and print all secrets path recoursive
- remove version in linter workflow
- remove gopass options
### Fixed
- coredump when calling --help and want no global options printed
- linter issues

## [v2.16.3 - 2025-12-25]
### Changed
- use Go1.25
- update dependencies
- change password profile functions
- replace obsolete bitnami/openldap docker test container and scripts
- use golangci-lint v2
- update password profiles

## [v2.16.0 - 2025-03-08]
### New
- config get key option
- config print/list/show output as json
- create missing directories
### Changed
- use Go1.24
- config file search and processing
- try to create directories for config, keys and data if not exists
- update dependencies
- update docker test container

# [v2.15.0 - 2025-02-27]
### New
- search case insensitive per default
- add case-sensitive option to get
- add list passwordprofiles to check and genpass
- new option password_profiles for genpass, checkpass and ldap
### Changed
- rename check command to checkpass and use check as alias
- use password profile functions from pwlib 1.10+
- update tests
- change search order for configs
- use separate file for profile sets
- allow more special chars
- enhance README`

## [v2.14.0 - 2025-02-21]
### New
- add profileset option for genpass/checkpass
- add list option for get command
### Changed
- update dependencies
### Fixed
- fix panic in vault list command

## [v2.13.0 - 2024-10-24]
### New
- add ldap members command
### Changed
- update dependencies

## [v2.12.1 - 2024-10-03]
### New
- add arm64 target
### Changed
- use Go1.23
- update dependencies
- use Goreleaser V2 and v6 GitHub Action

## [v2.12.0 - 2024-09-24]
### New
- add new command `pwcli decrypt`
### Changed
- update dependencies
- update docker test container
- update actions

## [v2.11.2 - 2024-08-18]
### Changed
- refactor testinit
- replace os.ReadFile and os.WriteFile
- update dependencies
### Fixed
- fix new linter issues

## [v2.11.1 - 2024-08-03]
### New
- add Argon2 hash method
- replace --hash-mod switch with similar sub commands
### Changed
- update dependencies


## [v2.10.0 - 2024-07-25]
### New
- add Basic Auth (hash) method
### Changed
- update dependencies

## [v2.9.1 - 2024-05-25]
### Changed
- use Go1.22
- update dependencies

## [v2.9.0 - 2024-05-24]
### New
- add ssha hash method
### Changed
- add prefix and test option to hash command
- update dependencies

## [v2.8.0 - 2024-03-17]
### New
- add method kms get,list and encrypt for Amazon KMS Service
### Changed
- update dependencies
- refactor genkey command
- refactor vault flags and tests

## [v2.7.1 - 2024-03-01]
### New
- add scram, md5 and bcrypt password hash command
### Changed
- use gomodules v1.11.5
- add sshpubkey attribute if not already there

## [v2.6.1 - 2024-02-27]
### New
- add new command `pwcli ldap setpass` to set ldap passwords
- add new command `pwcli ldap setssh` to set public ssh keys
- add new command `pwcli ldap show` to retrieve ldap attributes
- add new command `pwcli ldap groups` to list membership of groups
- use bitname ldap test container
- add scripts and docs to packages
### Changed
- use gomodules v1.11.3
- update dependencies
- move docker container to test/docker
- remove tools.go
### fixed
- linter issues

## [v2.5.0 - 2023-10-27]
### Changed
- use go 1.21
- use gomodules v1.10.0
- update testinit
- update workflow

## [v2.4.7 - 2023-08-09]
### New
- add new flag --unit-tests to redirect output for unit tests
### Changed
- use gomodules v1.9.3
- use common.CmdRun instead of cmdTest
- use common.CmdFlagChanged instead of cmd.Flags().Lookup().Changed

## [v2.4.6 - 2023-08-04]
### New
- add version test
### Changed
- move tests to there packages

## [v2.4.5 - 2023-08-01]
### New
- add no-color switch to explicit disable colored output
### Changed
- refactor config file handling

## [v2.4.0 - 2023-07-16]
### New
- add config save --filename and --force switch
### Changed
- use go 1.20 and update dependencies
- update gomodule version to 1.9.0
- use docker_helper for tests
- load config file from $HOME/etc or current dir
- use default method openssl

## [2.3.1 - 2023-05-19]
### Changed
- update gomodule version to 1.7.4
- change goreleaser version date string and changelog

## [2.3.0 - 2023-05-11]
### New
- add vault method to get command
- add vault test container
### Changed
- update gomodule version to 1.7.1

## [v2.2.0 - 2023-05-07]
### New
- add vault read/write functions
- add totp functions
- add deb and rpm packages via nfpm
### Changed
- update gomodule version to 1.7.0

## [v2.0.0 - 2023-04-20]
initial public release
