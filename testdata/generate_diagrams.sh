#!/usr/bin/env bash

set -eu -o pipefail

declare -r GED_BIN="${GED_BIN:-"bin/main"}"

function main() {
  local -r infile="${1:?missing infile}"
  local -r outdir="${2:?missing outdir}"

  if [[ ! -x "${GED_BIN}" ]]; then
    echo >&2 "# GED_BIN '${GED_BIN}' must be executable"
    return 1
  fi

  if [[ ! -d "${outdir}" ]]; then
    mkdir -pv "${outdir}"
  fi

  for output_format in svg png mermaid; do
    echo >&2 "# testing -output-format ${output_format}"

    for direction in TB TD BT RL LR; do
      "${GED_BIN}" draw -output-format "${output_format}" -direction "${direction}" -display-id=true \
        < "${infile}" \
        > "${outdir}/direction_${direction}-with_id.${output_format}"

      "${GED_BIN}" draw -output-format "${output_format}" -direction "${direction}" -display-id=false \
        < "${infile}" \
        > "${outdir}/direction_${direction}-without_id.${output_format}"
    done
  done

  for output_format in svg png; do
    echo >&2 "# testing -input-format=mermaid -output-format ${output_format}"

    for direction in TB TD BT RL LR; do
      "${GED_BIN}" draw -input-format=mermaid -output-format "${output_format}" -direction "${direction}" -display-id=true \
        < "${outdir}/direction_${direction}-with_id.mermaid" \
        > "${outdir}/from_mermaid-direction_${direction}-with_id.${output_format}"

      "${GED_BIN}" draw -input-format=mermaid -output-format "${output_format}" -direction "${direction}" -display-id=false \
        < "${outdir}/direction_${direction}-without_id.mermaid" \
        > "${outdir}/direction_${direction}-without_id.${output_format}"
    done
  done
}

main "${@}"
