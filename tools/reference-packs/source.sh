#!/usr/bin/env bash
set -euo pipefail

if [[ ! -x ./build/bin/reference-pack-producer ]]; then
  echo 'Run make reference-pack-producer-build before source preparation.' >&2
  exit 2
fi

producer_report="$(mktemp)"
trap 'rm -f -- "$producer_report"' EXIT
producer_status=0
./build/bin/reference-pack-producer \
  --action "${ACTION:-check}" \
  --lock "${SOURCE_LOCK:?explicit source input required}" \
  --directory "${SOURCE_DIR:?explicit source directory required}" \
  --built-at "${BUILT_AT:-}" \
  --editorial "${EDITORIAL:-}" \
  --input "${SOURCE_INPUT:-}" \
  --report-file "$producer_report" || producer_status=$?
if [[ "${ACTION:-check}" == audit && -s "$producer_report" ]]; then
  "${NODE_BIN:?}" ./tools/reference-packs/witness.mjs source-audit "${SOURCE_LOCK:?}" "$producer_report" "${SOURCE_DIR:?}"
fi
exit "$producer_status"
