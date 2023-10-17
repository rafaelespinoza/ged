GO := "go"
BIN_DIR := "bin"
MAIN_BIN := BIN_DIR / "main"

SRC_PATHS := ". ./internal/..."

# General-purpose arguments to pass to a command. May be overridden at invocation.
# Example: just ARGS='-foo=bar -v' <recipe_name>

ARGS := ""

# list available recipes
default:
    just --list --unstable --unsorted -f {{ justfile() }}

_bindir:
    @mkdir -pv {{ BIN_DIR }}

# compile a binary for package main
build: _bindir
    @{{ GO }} build -o {{ MAIN_BIN }} {{ invocation_directory() }}

alias b := build

# execute the built binary
run +ARGS:
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
