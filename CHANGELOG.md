# Changelog pwcli

## [v2.4.4 - 2023-07-27]
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
