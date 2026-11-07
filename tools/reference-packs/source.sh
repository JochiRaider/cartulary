#!/usr/bin/env bash
set -euo pipefail

if [[ ! -x ./build/bin/reference-pack-producer ]]; then
  echo 'Run make reference-pack-producer-build before source preparation.' >&2
  exit 2
fi

exec ./build/bin/reference-pack-producer \
  --action "${ACTION:-check}" \
  --lock "${SOURCE_LOCK:-contracts/reference-pack-producers/attack/source-lock.v20.1.json}" \
  --directory "${SOURCE_DIR:-.cartulary/reference-packs/attack/20.1}" \
  --built-at "${BUILT_AT:-}"
