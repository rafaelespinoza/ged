#!/usr/bin/env -S just -f

GO := "go"
MAIN_BIN := "bin/main"
SRC_PATHS := ". ./internal/..."

# General-purpose arguments to pass to a command. May be overridden at invocation.
# Example: just ARGS='-foo=bar -v' <recipe_name>

ARGS := ""

# list available recipes
default:
    @{{ justfile() }} --list --unsorted

# compile a binary for package main
build: _bin_dir
    @{{ GO }} build -o {{ MAIN_BIN }} {{ invocation_directory() }}

alias b := build

# execute the built binary
@run +ARGS='help':
    {{ MAIN_BIN }} {{ ARGS }}

alias r := run

# build a new binary, run it
buildrun +ARGS='help': build (run ARGS)

alias br := buildrun

# get module dependencies, tidy them up, vendor them
mod:
    {{ GO }} mod tidy
    {{ GO }} mod vendor

# examine source code for suspicious constructs
vet ARGS='':
    {{ GO }} vet {{ ARGS }} {{ SRC_PATHS }}

# run tests (override variable value ARGS to use test flags)
test PKG_PATH='./...':
    {{ GO }} test {{ PKG_PATH }} {{ ARGS }}

_bin_dir:
    @mkdir -pv {{ parent_directory(MAIN_BIN) }}

# Recipes for shell script maintenance rely on
# * shellcheck: https://github.com/koalaman/shellcheck
# * shfmt: https://github.com/mvdan/sh

SCRIPTS := `git ls-files | grep '\.sh$'`

# run shellcheck on scripts
lint-scripts:
    shellcheck {{ SCRIPTS }}

# run shfmt on shell scripts
fmt-scripts:
    shfmt -ci -d -s -sr {{ SCRIPTS }}

# run shfmt on shell scripts and overwrite files
fmtw-scripts:
    shfmt -ci -d -s -sr -w {{ SCRIPTS }}
