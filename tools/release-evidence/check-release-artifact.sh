#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -ne 2 ]]; then
  echo "usage: check-release-artifact.sh <artifact-label> <artifact-path>" >&2
  exit 2
fi

label="$1"
artifact_path="$2"

if [[ ! -e "$artifact_path" ]]; then
  printf '%s artifact missing: %s\n' "$label" "$artifact_path" >&2
  exit 1
fi

if [[ ! -f "$artifact_path" ]]; then
  printf '%s artifact is not a regular file: %s\n' "$label" "$artifact_path" >&2
  exit 1
fi

if [[ ! -s "$artifact_path" ]]; then
  printf '%s artifact is empty: %s\n' "$label" "$artifact_path" >&2
  exit 1
fi

case "$label" in
  "license report")
    "${NODE_BIN:-node}" ./tools/release-evidence/validate-license-report.mjs "$artifact_path"
    ;;
  "SBOM")
    "${NODE_BIN:-node}" ./tools/release-evidence/validate-release-sbom.mjs "$artifact_path"
    ;;
  *)
    printf 'unsupported release artifact label: %s\n' "$label" >&2
    exit 2
    ;;
esac

printf '%s artifact present: %s\n' "$label" "$artifact_path"
