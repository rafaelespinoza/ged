#!/usr/bin/env -S just -f

GO := "go"

[private]
MAIN_BIN := "bin/main"

PKG_PATH := "./..."

# list available recipes
default:
    @{{ justfile() }} --list --unsorted

# compile a binary for package main
[group('go')]
build *args: _bin_dir
    @{{ GO }} build -o {{ MAIN_BIN }} {{ args }} {{ invocation_directory() }}

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

# examine source code for suspicious constructs
[group('go')]
vet *args:
    {{ GO }} vet {{ args }} {{ PKG_PATH }}

# run tests (override variable value ARGS to use test flags)
[group('go')]
test *args:
    {{ GO }} test {{ args }} {{ PKG_PATH }}

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
