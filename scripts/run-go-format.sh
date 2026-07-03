#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: run-go-format.sh --check|--write" >&2
  exit 2
}

if [[ "$#" -ne 1 ]]; then
  usage
fi

mode="$1"
case "$mode" in
  --check | --write) ;;
  *) usage ;;
esac

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
GOFMT_BIN="${GOFMT:-gofmt}"

# shellcheck source=tools/harness/generated-artifacts/generated-artifacts.sh
# shellcheck disable=SC1091
source "$ROOT_DIR/tools/harness/generated-artifacts/generated-artifacts.sh"

cd "$ROOT_DIR"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "go format check must run inside a git work tree" >&2
  exit 1
fi

if ! command -v "$GOFMT_BIN" >/dev/null 2>&1; then
  echo "missing gofmt: install Go and ensure gofmt is on PATH, or set GOFMT" >&2
  exit 1
fi

declare -A seen=()
go_files=()

append_go_files() {
  local path
  while IFS= read -r -d '' path; do
    if [[ -n "${seen[$path]:-}" ]]; then
      continue
    fi
    seen[$path]=1
    if cartulary_is_authored_go_file "$path"; then
      go_files+=("$path")
    fi
  done
}

append_go_files < <(git ls-files -z -- '*.go')
append_go_files < <(git ls-files -z --others --exclude-standard -- '*.go')

if [[ "${#go_files[@]}" -eq 0 ]]; then
  exit 0
fi

case "$mode" in
  --check)
    unformatted="$("$GOFMT_BIN" -l -- "${go_files[@]}")"
    if [[ -n "$unformatted" ]]; then
      echo "gofmt required for authored Go files:" >&2
      printf '%s\n' "$unformatted" >&2
      echo "run make format to rewrite authored Go and frontend sources." >&2
      exit 1
    fi
    ;;
  --write)
    "$GOFMT_BIN" -w -- "${go_files[@]}"
    ;;
esac
