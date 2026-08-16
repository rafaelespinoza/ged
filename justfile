#!/usr/bin/env -S just -f

GO := "go"

PKG_PATH := "./..."

_MAIN_BIN := "bin/main"

_PKG := "github.com/rafaelespinoza/ged"

# list available recipes
default:
    @{{ justfile() }} --list --unsorted

# can set this variable to '-s' to make a statically-linked binary.
# You'll want to use this syntax when the value has a leading dash:
#     $ just _MORE_LDFLAGS="-s" build
# As opposed to this:
#     $ just --set _MORE_LDFLAGS "-s" build
_MORE_LDFLAGS := ''

# Figure out build metadata, which relies on git and date.
# If a command is not present then fallback to a default value.
_HAS_GIT := `command -v git >/dev/null 2>&1 && \
    git rev-parse --is-inside-work-tree >/dev/null 2>&1 && \
    echo "true" || \
    echo "false"`
_VERSION_BRANCH     := if _HAS_GIT == "true" { `git rev-parse --abbrev-ref HEAD` } else { "unknown" }
_VERSION_COMMIT     := if _HAS_GIT == "true" { `git rev-parse --short=7 HEAD` } else { "unknown" }
_VERSION_TAG        := if _HAS_GIT == "true" { `git describe --tag 2>/dev/null || echo "dev"` } else { "dev" }
_VERSION_BUILD_TIME := `command -v date >/dev/null 2>&1 && date -u +%FT%T%z || echo 'unknown'`

# compile a binary for package main
[group('go')]
[script('sh')]
build: _bin_dir
    set -eu
    _version_branch='{{ _VERSION_BRANCH }}'
    _version_build_time='{{ _VERSION_BUILD_TIME }}'
    _version_commit_hash='{{ _VERSION_COMMIT }}'
    _version_tag='{{ _VERSION_TAG }}'
    _go_version="$({{ GO }} version | awk '{ print $3 }')"
    _ld_delimiter_and_base_prefix="-X {{ _PKG }}/internal/cmd"
    ldflags="-extldflags '-static' \
      ${_ld_delimiter_and_base_prefix}.versionBranchName=${_version_branch} \
      ${_ld_delimiter_and_base_prefix}.versionBuildTime=${_version_build_time} \
      ${_ld_delimiter_and_base_prefix}.versionCommitHash=${_version_commit_hash} \
      ${_ld_delimiter_and_base_prefix}.versionTag=${_version_tag} \
      ${_ld_delimiter_and_base_prefix}.versionGoVersion=${_go_version} \
      {{ (_MORE_LDFLAGS) }}"
    bin='{{ clean(_MAIN_BIN) }}'
    mkdir -pv '{{ parent_directory(_MAIN_BIN) }}'
    {{ GO }} build -o="${bin}" -v -ldflags="${ldflags}" .
    "${bin}" version

alias b := build

# execute the built binary
[group('go')]
@run *args='help':
    {{ _MAIN_BIN }} {{ args }}

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
    @mkdir -pv {{ parent_directory(_MAIN_BIN) }}

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

CONTAINER_TOOL := 'podman'

# Here's a general but informal scheme for container image naming.
#     [REGISTRY[:PORT]/][NAMESPACE/]REPOSITORY[:TAG][@DIGEST]
# source: https://specs.opencontainers.org/distribution-spec/

_IMG_REGISTRY := 'localhost' # Consider changing to something more public at some point.
_IMG_NAMESPACE := 'rafaelespinoza'
_IMG_REPOSITORY := 'ged'
_IMG_FULL_NAME := _IMG_REGISTRY / _IMG_NAMESPACE / _IMG_REPOSITORY
_SRC_URL := 'https://github.com/rafaelespinoza/ged'

# build a container image for the explore-data script
[group('container')]
[script('sh')]
[positional-arguments]
build-container-img *args:
    set -eu

    _usage () {
        printf >&2 'Usage: just build-container-img [-c] [-h]

    Build a container image.
    By default, the build context is from this directory.
    To build from a fresh clone, pass the flag -c.

    Settings:
        Image name: "{{ _IMG_FULL_NAME }}"
        Default build context: "{{ justfile_directory() }}"

    Flags:
        -c
            clone repo
        -h
            show help\n'
    }

    # Setup flag parsing
    _clone_repo=0
    OPTIND=1
    while getopts :ch opt; do
        case "${opt}" in
            c) _clone_repo=1 ;;
            h) _usage; exit 0;;
            \?) _usage; exit 1;;
        esac
    done
    shift $((OPTIND - 1))

    umask 077
    build_dir='{{ justfile_directory() }}'
    if [ "${_clone_repo}" -eq 1 ]; then
        build_dir=$(mktemp -d -p /tmp {{ _IMG_REPOSITORY }}.img_build.XXXXXX)
        trap 'rm -rf -- "${build_dir}"' EXIT
        git clone --depth=1 'file://{{ justfile_directory() }}' "${build_dir}"
    fi
    {{ CONTAINER_TOOL }} image build \
        --build-arg VERSION_BRANCH='{{ _VERSION_BRANCH }}' \
        --build-arg VERSION_COMMIT='{{ _VERSION_COMMIT }}' \
        --build-arg VERSION_TAG='{{ _VERSION_TAG }}' \
        --label 'org.opencontainers.image.url='{{ _SRC_URL }} \
        --tag '{{ _IMG_FULL_NAME }}' \
        --file Containerfile "${build_dir}"

# remove builder images, dangling runtime images; interactive prompt
[group('container')]
[script('sh')]
prune-container-imgs:
    set -eu

    builder_img_label='com.rafaelespinoza.ged.stage_type=builder'
    echo >&2 "# targeting builder image stage, --filter label=${builder_img_label}"
    # show what would be pruned
    {{ CONTAINER_TOOL }} image ls --filter "label=${builder_img_label}"
    # prompt for confirmation before pruning
    {{ CONTAINER_TOOL }} image prune --filter "label=${builder_img_label}"

    runtime_img_label='org.opencontainers.image.url={{ _SRC_URL }}'
    echo >&2 "# targeting runtime image stage, --filter label=${runtime_img_label} --filter dangling=true"
    # show what would be pruned
    {{ CONTAINER_TOOL }} image ls --filter "label=${runtime_img_label}" --filter dangling=true
    # prompt for confirmation before pruning
    {{ CONTAINER_TOOL }} image prune --filter "label=${runtime_img_label}" --filter dangling=true
