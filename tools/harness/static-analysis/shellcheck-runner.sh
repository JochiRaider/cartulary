#!/usr/bin/env bash
set -euo pipefail

: "${CARTULARY_SHELLCHECK_FILE_LIST:?}"
: "${CARTULARY_STATIC_CACHE_STAMP:?}"
: "${SHELLCHECK_BIN:?}"

shell_files=()
mapfile -d '' -t shell_files <"$CARTULARY_SHELLCHECK_FILE_LIST"

"$SHELLCHECK_BIN" "${shell_files[@]}"

mkdir -p "$(dirname "$CARTULARY_STATIC_CACHE_STAMP")"
printf 'lint-shell ok\n' >"$CARTULARY_STATIC_CACHE_STAMP"
