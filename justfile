#!/usr/bin/env -S just -f

GO := "go"
MAIN_BIN := "bin/main"

SRC_PATHS := ". ./internal/..."

# General-purpose arguments to pass to a command. May be overridden at invocation.
# Example: just ARGS='-foo=bar -v' <recipe_name>

ARGS := ""

RUN_DIR := "/tmp/ged"

# list available recipes
default:
    @{{ justfile() }} --list --unstable --unsorted

# compile a binary for package main
build: _bin_dir
    @{{ GO }} build -o {{ MAIN_BIN }} {{ invocation_directory() }}

alias b := build

# execute the built binary
run +ARGS: _run_dir
    {{ MAIN_BIN }} {{ ARGS }}

alias r := run

# build a new binary, run it
buildrun +ARGS: build (run ARGS)

alias br := buildrun

# get dependencies and vendor them
deps:
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

_run_dir:
    @mkdir -pv {{ RUN_DIR }} && chmod 700 {{ RUN_DIR }}
