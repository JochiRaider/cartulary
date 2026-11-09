#!/usr/bin/env bash
set -euo pipefail

if [[ ! -x "./build/bin/reference-pack-${PRODUCER:?}" ]]; then
  echo 'Run make reference-pack-producer-build PRODUCER=<pack> before source preparation.' >&2
  exit 2
fi

audit_work="$(mktemp -d)"
trap 'rm -rf -- "$audit_work"' EXIT
producer_report="$audit_work/report.json"
# The delivery file exists even when a producer emits no report.
: > "$producer_report"
context_status=0
if [[ "${ACTION:-check}" == audit ]]; then
  "${NODE_BIN:?}" ./tools/reference-packs/audit-report.mjs context "$PRODUCER" \
    "${SOURCE_LOCK:?}" "./build/bin/reference-pack-${PRODUCER}" "${BUILT_AT:-}" "${EDITORIAL:-}" "${SOURCE_DIR:?}" > "$audit_work/context.json" 2> "$audit_work/context-error.txt" || context_status=$?
fi
producer_status=0
"./build/bin/reference-pack-${PRODUCER:?}" \
  --action "${ACTION:-check}" \
  --lock "${SOURCE_LOCK:?explicit source input required}" \
  --directory "${SOURCE_DIR:?explicit source directory required}" \
  --built-at "${BUILT_AT:-}" \
  --editorial "${EDITORIAL:-}" \
  --input "${SOURCE_INPUT:-}" \
  --report-file "$producer_report" || producer_status=$?
# A producer failure is primary; its diagnostics/report have already been
# delivered. Never turn it into success or mask it with a secondary witness error.
if (( producer_status != 0 )); then
  exit "$producer_status"
fi
if [[ "${ACTION:-check}" == audit ]]; then
  if (( context_status != 0 )); then
    cat -- "$audit_work/context-error.txt" >&2
    exit "$context_status"
  fi
  "${NODE_BIN:?}" ./tools/reference-packs/audit-report.mjs validate "$PRODUCER" "$producer_report" "$audit_work/context.json"
  if [[ "$PRODUCER" != attack ]]; then
    "$NODE_BIN" ./tools/reference-packs/source-witness.mjs source-audit "${SOURCE_LOCK:?}" "$producer_report" "${SOURCE_DIR:?}" "$audit_work/context.json"
  fi
fi
