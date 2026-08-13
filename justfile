#!/usr/bin/env -S just -f

GO := "go"

[private]
MAIN_BIN := "bin/main"

PKG_PATH := "./..."

[private]
_GO_VERSION := `go version | awk '{ print $3 }'`
[private]
PKG_IMPORT_PATH := "github.com/rafaelespinoza/ged"
[private]
_LDFLAGS_BASE_PREFIX := "-X " + PKG_IMPORT_PATH + "/internal/cmd"
[private]
_LDFLAGS_DELIMITER := "\n\t"
[private]
_LDFLAGS := ("-extldflags '-static'" + _LDFLAGS_DELIMITER + _LDFLAGS_BASE_PREFIX + ".versionBranchName=" + `git rev-parse --abbrev-ref HEAD` + _LDFLAGS_DELIMITER + _LDFLAGS_BASE_PREFIX + ".versionBuildTime=" + `date -u +%FT%T%z` + _LDFLAGS_DELIMITER + _LDFLAGS_BASE_PREFIX + ".versionCommitHash=" + `git rev-parse --short=7 HEAD` + _LDFLAGS_DELIMITER + _LDFLAGS_BASE_PREFIX + ".versionGoVersion=" + _GO_VERSION + _LDFLAGS_DELIMITER + _LDFLAGS_BASE_PREFIX + ".versionTag=" + `git describe --tag 2>/dev/null || echo 'dev'`)

# list available recipes
default:
    @{{ justfile() }} --list --unsorted

# compile a binary for package main
[group('go')]
[script('sh')]
build *args: _bin_dir
    set -eu
    bin={{ clean(MAIN_BIN) }}
    bin_dir={{ parent_directory(MAIN_BIN) }}
    mkdir -pv "${bin_dir}"
    ldflags="{{ _LDFLAGS }}"
    {{ GO }} build -o="${bin}" -v -ldflags="${ldflags}" {{ args }} .
    "${bin}" version

alias b := build

# execute the built binary
[group('go')]
@run *args='help':
    {{ MAIN_BIN }} {{ args }}

alias r := run

# build a new binary, run it
[group('go')]
buildrun *runArgs='help': build (run runArgs)

alias br := buildrun

# get module dependencies, tidy them up
[group('go')]
modtidy:
    {{ GO }} mod tidy

# run tests
[group('go')]
test *args:
    {{ GO }} test {{ PKG_PATH }} {{ args }}

# examine source code for suspicious constructs
[group('go.static')]
vet *args:
    {{ GO }} vet {{ args }} {{ PKG_PATH }}

# check for known vulnerabilities
[group('go.static')]
govulncheck *args:
    {{ GO }} run golang.org/x/vuln/cmd/govulncheck@latest {{ args }} {{ PKG_PATH }}

GOSEC := "gosec"

# This Justfile won't install the scanner binary for you, so check out the
# gosec README for instructions: https://github.com/securego/gosec
#
# If necessary, specify the path to the built binary with the GOSEC variable.

# Run a security scanner over the source code
[group('go.static')]
gosec *args:
    {{ GOSEC }} {{ args }} {{ PKG_PATH }}

_bin_dir:
    @mkdir -pv {{ parent_directory(MAIN_BIN) }}

# Recipes for shell script maintenance rely on
# * shellcheck: https://github.com/koalaman/shellcheck
# * shfmt: https://github.com/mvdan/sh

SCRIPTS := `git ls-files | grep '\.sh$'`

# run shellcheck on scripts
[group('scripts')]
lint-scripts:
    shellcheck {{ SCRIPTS }}

# run shfmt on shell scripts
[group('scripts')]
fmt-scripts:
    shfmt -ci -d -s -sr {{ SCRIPTS }}

# run shfmt on shell scripts and overwrite files
[group('scripts')]
fmtw-scripts:
    shfmt -ci -d -s -sr -w {{ SCRIPTS }}
