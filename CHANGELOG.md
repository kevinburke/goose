# Changelog

All notable changes to this project are documented here.

This repository did not previously have a changelog. Per project policy, this
file summarizes the last five years of tagged releases based on the git
history. Earlier tags exist, but releases before `1.15` are outside that
window.

## 1.17.0 - 2026-04-11

- Switched the Postgres driver from `lib/pq` to `github.com/jackc/pgx/v5`.
- Added programmatic `goosedb.DBConf` constructors for callers that want to
  configure goose without writing `dbconf.yml` to disk.

## 1.16.0 - 2026-04-11

- Turned `lib/goosedb` warnings into errors.
- Updated test coverage for the latest Go toolchain and fixed vet failures in
  `cmd/goose`.
- Added `go.mod` and refreshed vendored dependencies to match module-based
  builds.
- Updated installation guidance to prefer `go install` over `go get`.
- Refreshed dependencies and CI configuration for 2026.
- Aligned command release versioning with the broader release workflow.

## 1.15 - 2021-05-03

- Updated vendored database drivers, including `lib/pq`.
- Added GitHub Actions CI.
- Simplified Makefile package list handling.

## Thanks

See the project's earlier contributor list in
[README.md](/Users/kevin/src/github.com/kevinburke/goose/worktrees/changelog/README.md).
