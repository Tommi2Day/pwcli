# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Lint (golangci-lint v2 required)
golangci-lint run ./...

# Run all tests (unit only, no Docker)
SKIP_KMS=1 SKIP_VAULT=1 SKIP_LDAP=1 SKIP_AWSSM=1 go test ./cmd/ -v -count=1

# Run a specific test function
go test ./cmd/ -run TestGopassCLI -v -count=1

# Run with Docker integration tests (requires Docker)
go test ./cmd/ -v -count=1 -timeout 120s
```

Skip environment variables for Docker-backed tests: `SKIP_KMS`, `SKIP_VAULT`, `SKIP_LDAP`, `SKIP_AWSSM`.

## Architecture

### Entry point and command structure

`main.go` calls `cmd.Execute()`. Every command lives in the `cmd/` package and registers itself to `RootCmd` (defined in `cmd/rootcmd.go`) via `init()`. The package-level variable `pc *pwlib.PassConfig` is the central config object created in `initConfig()` and shared by all commands.

### Dependency on gomodules (pwlib)

The heavy lifting — all encryption, decryption, and password store operations — is delegated to the library `github.com/tommi2day/gomodules/pwlib` (vendored under `vendor/github.com/tommi2day/gomodules/pwlib/`). The `cmd/` layer is thin: it parses flags, populates `pc`, and calls `pwlib` functions. When adding new crypto or store functionality, check what `pwlib` already provides before writing native code.

The `common` package (`github.com/tommi2day/gomodules/common`) provides shared utilities: `CmdRun` for test execution, `WriteStringToFile`, `GetDockerPool`, `CmdFlagChanged`, `PromptPassword`.

AWS access (KMS, Secrets Manager, RDS) also goes through `pwlib`: `configureAWS()` in `rootcmd.go` applies `--aws_profile` / `--aws_mfa_token` via `pwlib.SetAWSProfile*`. `pwlib.ConnectToKMS` returns an error rather than exiting.

### Configuration

Configuration is managed by cobra + viper. The config file defaults to `pwcli.yaml` or `<app>.yaml`, searched in `.`, `$HOME/.pwcli`, `$HOME/etc`, `/etc/pwcli`. The `--method` flag selects the encryption/store backend. Valid methods are defined as constants in `rootcmd.go`: `openssl` (default), `go`, `enc`, `plain`, `age`, `gpg`, `vault`, `kms`, `gopass`, `awssm`, `rds`.

Commands that talk to AWS: `kms`, `awssm` (`read`/`write`/`secrets`, one file `cmd/awssm.go`) and `rds` (`rds token`, `cmd/rds.go`). `get` supports them as `--method kms|awssm|rds`. `vault read` and `awssm read` share `printSecretData`/`printAllData` for the `--json`, `--export` and `--dotenv` output formats.

### Password file format (local store)

Plaintext files use colon-delimited `system:user:password` lines. The special system name `!default` matches any system. These files are encrypted to produce `.gp` (go method) or `.pw` (openssl) files stored in `datadir`.

### Test patterns

**Unit tests** (no external dependencies): `pwcli_test.go`, `gopass_test.go`, `hash_test.go`, etc.

**Docker integration tests**: files named `*_docker_test.go` spin up containers via `ory/dockertest/v4`. They check the `SKIP_*` env vars at startup and skip if set. KMS and Secrets Manager tests run against `motoserver/moto` containers; the shared launch helper is `runMotoContainer` in `cmd/moto_docker_test.go`. The LDAP container declares no `EXPOSE` ports, so `ldap_docker_test.go` binds 389/636 explicitly.

**Capturing output in tests**: all assertions go against `common.CmdRun(RootCmd, args)` which captures what is written to cobra's output buffer (`cmd.Printf` / `cmd.Println`). Always pass `--unit-test` to redirect logrus to that same buffer. The `get` command's final password is printed via `fmt.Println` (bypasses capture); those tests assert on the log message `"Found matching entry: 'value'"` using `--info --unit-test`.

**Cobra flag state leaks between subtests**: cobra/viper retain flag values across `Execute()` calls in the same process. Call `viper.Reset()` and reset package-level vars (e.g. `gopassIdentityDir = ""`) between subtests that would otherwise inherit state.

**Test working directory**: `test/testinit.go` sets `test.TestDir` and `test.TestData` (`test/testdata/`) using `runtime.Caller(0)` and `os.Chdir`. The `testdata/` directory is gitignored; tests write all artifacts there.

### Linter notes

- `goimports` enforces formatting — var blocks must be consistently aligned.
- `gosec` G703 fires on `os.WriteFile`/`os.ReadFile` with paths built from package-level variables. Fix with `filepath.Clean(path)` where possible; fall back to `//nolint:gosec` for test-only paths that are provably safe.
- `revive` `unhandled-error` is configured to flag `fmt.Printf/Println/Fprintf` — use `//nolint:revive` or assign the return value where needed.
- Line length limit is 200 characters (`lll`).
- Cognitive and cyclomatic complexity thresholds are both 15 (`gocognit`, `gocyclo`).
