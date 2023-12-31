#!/usr/bin/env bash

set -eu -o pipefail

declare CURR_SCRIPT SCRIPT_DIR GED_BIN
CURR_SCRIPT="$(realpath "${0}")"
SCRIPT_DIR=$(dirname "${CURR_SCRIPT}")
GED_BIN="${GED_BIN:-"${SCRIPT_DIR}/bin/main"}"
readonly CURR_SCRIPT SCRIPT_DIR GED_BIN
export CURR_SCRIPT SCRIPT_DIR GED_BIN

declare -r fzf_url='https://github.com/junegunn/fzf'
declare -r bkt_url='https://github.com/dimo414/bkt'
declare -r err_no_fzf="Install fzf (${fzf_url}) to use a fuzzy finder on the input data. If you don't have fzf or don't want to use it, invoke the subcommand directly. You must know the person ID(s) ahead of time. An ID should be a GEDCOM Xref for an IndividualRecord, ie: '@I123@'."
declare -r no_bkt_msg="Optionally install bkt (${bkt_url}) to enable caching."

# _run_fzf executes fzf with some common flags, so that the user experience is more consistent. It
# expects its standard input to be the output of the ged subcommand "parse to-lines". Callers of
# this func should further specify any fzf flags as positional args.
function _run_fzf() {
	# The fields in the output of "parse to-lines" is delimited by this character.
	local -r delimiter='\t'

	# Omit display of first field, ID, b/c it's a long and generated value. But do use that ID value
	# in looking up the data.
	local -r with_nth='2..'

	# Prefer search results where the principle person's name, rather than the name of a parent, is
	# a match.
	local -r tiebreak=begin

	fzf -d "${delimiter}" \
		--with-nth="${with_nth}" \
		--info=inline \
		--border=bold --margin=2 --padding=2 \
		--preview-window='down,90%,border-horizontal,wrap' \
		--bind='ctrl-space:change-preview-window(70%|50%|hidden|)' \
		--bind='change:first' \
		--tiebreak="${tiebreak}" \
		--layout=reverse \
		--cycle \
		"$@"
}

function _run_ged() {
	if [[ ! -x ${GED_BIN} ]]; then
		echo >&2 "# it appears that GED_BIN (${GED_BIN}) is not executable"
		return 1
	fi

	"${GED_BIN}" -q "${@}"
}

function relate() {
	if ! command -v fzf >&/dev/null; then
		echo "${err_no_fzf}

Example:
$ bin/main explore-data relate -p1 @I0@ -p2 @I10@ < testdata/kennedy.ged" | fold -s >&2
		return 1
	fi

	local -r infile="${1:?missing infile}"
	if [[ ! -r ${infile} ]]; then
		echo >&2 "file not found or not readable"
		return 1
	fi

	# Allow for the preview script to read in the data it needs.
	export INFILE="${infile}"

	# This little preview script is intepreted by fzf. NOTE: it won't have access to any variables
	# created in the encompassing script unless they are exported or are made explicitly available
	# to the invocation of fzf.
	#
	# shellcheck disable=SC2016
	local fzf_preview='${GED_BIN:?missing GED_BIN} -q explore-data show --target-id "$(echo {1})" <"${INFILE:?missing INFILE}"'
	if command -v bkt >&/dev/null; then
		fzf_preview="bkt -- ${fzf_preview}"
	else
		echo >&2 "${no_bkt_msg}"
	fi

	# The same environmental constraints apply to this script.
	# Use expression "{+1}" to get all of the selected lines as space-separated values (this is the
	# "+") and pick the 1st field (this is "1") of each. The 1st field is the Xref of the selected
	# person. That format is determined by the ged subcommand, parse to-lines.
	#
	# shellcheck disable=SC2016
	local -r fzf_on_enter='
	p1=$(echo {+1} | awk "{ print \$1 }")
	p2=$(echo {+1} | awk "{ print \$2 }")
	${GED_BIN:?missing GED_BIN} -q explore-data relate -p1 "${p1}" -p2 "${p2}" <"${INFILE:?missing INFILE}"'

	_run_ged parse to-lines < "${infile}" | _run_fzf \
		--header='Pick 2 people to compare. Press TAB to select/unselect. Press Enter when ready.' \
		--multi=2 \
		--preview="${fzf_preview}" \
		--bind="enter:become(${fzf_on_enter})"
}

function show() {
	if ! command -v fzf >&/dev/null; then
		echo "${err_no_fzf}

Example:
$ bin/main explore-data show --target-id @I10@ < testdata/kennedy.ged" | fold -s >&2
		return 1
	fi

	local -r infile="${1:?missing infile}"
	if [[ ! -r ${infile} ]]; then
		echo >&2 "file not found or not readable"
		return 1
	fi

	# Allow for the preview script to read in the data it needs.
	export INFILE="${infile}"

	# This little preview script is intepreted by fzf. NOTE: it won't have access to any variables
	# created in the encompassing script unless they are exported or are made explicitly available
	# to the invocation of fzf.
	#
	# shellcheck disable=SC2016
	local fzf_preview='${GED_BIN:?missing GED_BIN} -q explore-data show --target-id "$(echo {1})" <"${INFILE:?missing INFILE}"'
	if command -v bkt >&/dev/null; then
		fzf_preview="bkt -- ${fzf_preview}"
	else
		echo >&2 "${no_bkt_msg}"
	fi

	_run_ged parse to-lines < "${infile}" | _run_fzf \
		--header="Type in name of person to view" \
		--preview="${fzf_preview}" \
		--bind="enter:become(${fzf_preview})"
}

function usage() {
	echo "Usage: <action> [arguments]

Description:
	Wrapper script for ged to present GEDCOM data in a more human-friendly manner.

Dependencies:
	Requires ged, see the README for how to build. TLDR: $ just build
	Requires fzf (${fzf_url}).
	${no_bkt_msg}

	Requires GEDCOM-formatted data on your filesystem.
	See samples in repository directory, testdata/.

Actions:
	relate path_to_input.ged
		Calculate a relationship between 2 people.

	show path_to_input.ged
		Display some basic info about a person.
"
}

function main() {
	if [[ ${#} -lt 1 ]]; then
		usage
		return 0
	fi

	local -r action="${1:-help}"
	shift

	case "${action}" in
		relate)
			relate "${@}"
			;;
		show | view)
			show "${@}"
			;;
		*)
			usage
			;;
	esac
}

main "$@"
