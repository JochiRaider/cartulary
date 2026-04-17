#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
GENERATED_DIRS=(
  "internal/gen/contracts"
  "internal/gen/sql"
  "packages/protocol-ts/src/generated"
)

cd "$ROOT_DIR"

snapshot_generated_state() {
  local dir
  local -a files=()
  for dir in "${GENERATED_DIRS[@]}"; do
    if [[ -d "$dir" ]]; then
      while IFS= read -r file; do
        files+=("$file")
      done < <(find "$dir" -type f | LC_ALL=C sort)
    fi
  done

  if [[ "${#files[@]}" -eq 0 ]]; then
    return 0
  fi

  sha256sum "${files[@]}"
}

before_state="$(snapshot_generated_state)"
make generate
after_state="$(snapshot_generated_state)"

if [[ "$before_state" != "$after_state" ]]; then
  echo "generated artifact drift detected after make generate" >&2
  diff -u <(printf '%s\n' "$before_state") <(printf '%s\n' "$after_state") || true
  exit 1
fi
