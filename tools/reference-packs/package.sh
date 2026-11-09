#!/usr/bin/env bash
set -euo pipefail

if [[ ! -x "./build/bin/reference-pack-${PRODUCER:?}" ]]; then
  echo 'Run make reference-pack-producer-build PRODUCER=<pack> before package production.' >&2
  exit 2
fi
producer_report="$(mktemp)"
producer_work="$(mktemp -d)"
trap 'rm -f -- "$producer_report"; rm -rf -- "$producer_work"' EXIT
"${NODE_BIN:?}" ./tools/reference-packs/explicit-input.mjs "${INPUT:?}" "$producer_work/input.json"
cp -- "./build/bin/reference-pack-${PRODUCER:?}" "$producer_work/producer"
"$NODE_BIN" tools/reference-packs/isolated-package.mjs "$producer_work/producer" "$producer_work/input.json" "${ACTION:?}" "$producer_work" > "$producer_report"
"${NODE_BIN:?}" ./tools/reference-packs/dataset-witness.mjs "$ACTION" "$INPUT" "$producer_report"
