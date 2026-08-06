#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
GO_TMP_DIR="${GO_TMP_DIR:?GO_TMP_DIR is required}"
GOVULNCHECK_BIN="${GOVULNCHECK_BIN:-$ROOT_DIR/tmp/toolbin/govulncheck-v1.3.0}"
GOVULNCHECK_FLAGS="${GOVULNCHECK_FLAGS:--test -json}"
GOVULNCHECK_PATTERNS="${GOVULNCHECK_PATTERNS:-./cmd/... ./internal/... ./db/... ./tools/...}"
GOVULNCHECK_DB="${GOVULNCHECK_DB:-}"
NODE_BIN="${NODE_BIN:-}"

# shellcheck source=tools/harness/generated-artifacts/generated-artifacts.sh
# shellcheck disable=SC1091
source "$ROOT_DIR/tools/harness/generated-artifacts/generated-artifacts.sh"

resolve_node_bin() {
  if [[ -n "$NODE_BIN" && -x "$NODE_BIN" ]]; then
    printf '%s\n' "$NODE_BIN"
    return 0
  fi
  if [[ -x "$ROOT_DIR/tmp/node-runtime/bin/node" ]]; then
    printf '%s\n' "$ROOT_DIR/tmp/node-runtime/bin/node"
    return 0
  fi
  if command -v node >/dev/null 2>&1; then
    command -v node
    return 0
  fi
  return 1
}

if [[ "$GO_BIN" != */* ]] && command -v "$GO_BIN" >/dev/null 2>&1; then
  GO_BIN="$(command -v "$GO_BIN")"
elif [[ "$GO_BIN" != /* ]]; then
  GO_BIN="$ROOT_DIR/$GO_BIN"
fi

if [[ ! -x "$GO_BIN" ]]; then
  echo "go-vulncheck requires an executable GO at $GO_BIN" >&2
  exit 1
fi

if [[ "$GOVULNCHECK_BIN" != */* ]] && command -v "$GOVULNCHECK_BIN" >/dev/null 2>&1; then
  GOVULNCHECK_BIN="$(command -v "$GOVULNCHECK_BIN")"
elif [[ "$GOVULNCHECK_BIN" != /* ]]; then
  GOVULNCHECK_BIN="$ROOT_DIR/$GOVULNCHECK_BIN"
fi

if [[ ! -x "$GOVULNCHECK_BIN" ]]; then
  echo "go-vulncheck requires an executable GOVULNCHECK_BIN at $GOVULNCHECK_BIN" >&2
  echo "run make go-security-toolchain before go-vulncheck or set GOVULNCHECK_BIN to a ready govulncheck binary" >&2
  exit 1
fi

cd "$ROOT_DIR"

args=()
if [[ -n "$GOVULNCHECK_DB" ]]; then
  args+=("-db" "$GOVULNCHECK_DB")
fi
if [[ -n "$GOVULNCHECK_FLAGS" ]]; then
  read -r -a flag_args <<<"$GOVULNCHECK_FLAGS"
  args+=("${flag_args[@]}")
fi
if [[ -z "$GOVULNCHECK_PATTERNS" ]]; then
  echo "go-vulncheck requires at least one GOVULNCHECK_PATTERNS entry" >&2
  exit 1
fi
read -r -a patterns <<<"$GOVULNCHECK_PATTERNS"

mapfile -t packages < <(
  GOCACHE="$GO_CACHE_DIR" \
  GOMODCACHE="$GO_MOD_CACHE_DIR" \
  GOTMPDIR="$GO_TMP_DIR" \
    "$GO_BIN" list "${patterns[@]}" |
    cartulary_filter_authored_go_packages
)

if [[ "${#packages[@]}" -eq 0 ]]; then
  echo "go-vulncheck package discovery returned no authored packages" >&2
  exit 1
fi

tmp_dir=""
if [[ -n "${CARTULARY_STEP_ARTIFACT_DIR:-}" ]]; then
  mkdir -p "$CARTULARY_STEP_ARTIFACT_DIR"
  raw_output="$CARTULARY_STEP_ARTIFACT_DIR/govulncheck-output.jsonstream"
  findings_output="$CARTULARY_STEP_ARTIFACT_DIR/govulncheck-findings.json"
else
  tmp_dir="$(mktemp -d)"
  raw_output="$tmp_dir/govulncheck-output.jsonstream"
  findings_output="$tmp_dir/govulncheck-findings.json"
fi

trap 'if [[ -n "$tmp_dir" ]]; then rm -rf "$tmp_dir"; fi' EXIT

set +e
env GOCACHE="$GO_CACHE_DIR" \
  GOMODCACHE="$GO_MOD_CACHE_DIR" \
  GOTMPDIR="$GO_TMP_DIR" \
  PATH="$(dirname "$GO_BIN"):$PATH" \
  "$GOVULNCHECK_BIN" "${args[@]}" "${packages[@]}" >"$raw_output"
scan_status=$?
set -e

cat "$raw_output"

node_bin="$(resolve_node_bin || true)"
if [[ -z "$node_bin" ]]; then
  echo "go-vulncheck requires node to parse Govulncheck JSON findings" >&2
  if [[ "$scan_status" -ne 0 ]]; then
    exit "$scan_status"
  fi
  exit 1
fi

set +e
"$node_bin" "$ROOT_DIR/tools/harness/static-analysis/govulncheck-findings.mjs" \
  --input "$raw_output" \
  --output "$findings_output"
findings_status=$?
set -e

case "$findings_status" in
  0)
    if [[ "$scan_status" -ne 0 ]]; then
      exit "$scan_status"
    fi
    exit 0
    ;;
  1)
    exit 1
    ;;
  *)
    if [[ "$scan_status" -ne 0 ]]; then
      exit "$scan_status"
    fi
    exit 1
    ;;
esac
