#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target_file="${repo_root}/pkg/discovery/oui/oui.csv"
tmp_file="$(mktemp)"
trap 'rm -f "${tmp_file}"' EXIT

url="https://standards-oui.ieee.org/oui/oui.csv"
user_agent="whosthere-oui-updater/1.0 (+https://github.com/ramonvermeulen/whosthere)"
expected_header="Registry,Assignment,Organization Name,Organization Address"

curl --fail --silent --show-error --location \
  --header "Accept: text/csv,application/vnd.ms-excel;q=0.9,*/*;q=0.8" \
  --user-agent "${user_agent}" \
  --output "${tmp_file}" \
  "${url}"

if [[ ! -s "${tmp_file}" ]]; then
  echo "downloaded OUI CSV is empty" >&2
  exit 1
fi

header="$(head -n 1 "${tmp_file}" | tr -d '\r')"
if [[ "${header}" != "${expected_header}" ]]; then
  echo "unexpected OUI CSV header: ${header}" >&2
  exit 1
fi

if cmp -s "${tmp_file}" "${target_file}"; then
  echo "OUI table is already up to date"
  exit 0
fi

mv "${tmp_file}" "${target_file}"
echo "Updated ${target_file}"
