# Changelog pwcli

## [v2.6.1 - 2024-02-16]
### New
- add new command `pwcli ldap setpass` to set ldap passwords
- add new command `pwcli ldap setssh` to set public ssh keys
- add new command `pwcli ldap show` to retrieve ldap attributes
- add new command `pwcli ldap groups` to list membership of groups
- use bitname ldap test container
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
