#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -eq 0 ]]; then
  printf 'usage: %s IMAGE [IMAGE...]\n' "$0" >&2
  exit 2
fi

for dependency in curl sha256sum tar; do
  command -v "${dependency}" >/dev/null || {
    printf '%s is required\n' "${dependency}" >&2
    exit 127
  }
done

grype_version="0.112.0"
grype_checksum="acb14a030010fe9bdb9594b4ae108d9d14ef2f926d936aa0916dc62c89c058ea"
tool_dir="${RUNNER_TEMP:-/tmp}/statesight-tools"
archive="${tool_dir}/grype.tar.gz"
grype="${tool_dir}/grype"

mkdir -p "${tool_dir}"
if [[ ! -x "${grype}" ]]; then
  curl --fail --location --retry 3 --output "${archive}" \
    "https://github.com/anchore/grype/releases/download/v${grype_version}/grype_${grype_version}_linux_amd64.tar.gz"
  printf '%s  %s\n' "${grype_checksum}" "${archive}" | sha256sum --check -
  tar --extract --gzip --file "${archive}" --directory "${tool_dir}" grype
fi

for image in "$@"; do
  "${grype}" "${image}" --only-fixed --fail-on high --output table
done
