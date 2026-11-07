#!/usr/bin/env bash
set -euo pipefail

if [[ ! -x ./build/bin/reference-pack-producer ]]; then
  echo 'Run make reference-pack-producer-build before package production.' >&2
  exit 2
fi
producer_report="$(mktemp)"
trap 'rm -f -- "$producer_report"' EXIT
./build/bin/reference-pack-producer --action "${ACTION:?}" --input "${INPUT:?}" > "$producer_report"
"${NODE_BIN:?}" ./tools/reference-packs/witness.mjs "$ACTION" "$INPUT" "$producer_report"
