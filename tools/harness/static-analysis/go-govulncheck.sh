#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/workspace_layout.generated.sh
source "$(dirname -- "${BASH_SOURCE[0]}")/../../workspace_layout.generated.sh"

# Raw scanner output and its parsed findings are retained security evidence. Keep
# their creation private even when this wrapper is invoked below an ambient 022
# umask; source generation does not occur in this process.
umask 077

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
GO_TMP_DIR="${GO_TMP_DIR:?GO_TMP_DIR is required}"
GOVULNCHECK_BIN="${GOVULNCHECK_BIN:-$ROOT_DIR/${CARTULARY_LAYOUT_TOOLBIN}/govulncheck-v1.3.0}"
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
  if [[ -x "$ROOT_DIR/${CARTULARY_LAYOUT_NODE_RUNTIME}/bin/node" ]]; then
    printf '%s\n' "$ROOT_DIR/${CARTULARY_LAYOUT_NODE_RUNTIME}/bin/node"
    return 0
  fi
  if command -v node >/dev/null 2>&1; then
    command -v node
    return 0
  fi
  return 1
}

node_bin="$(resolve_node_bin || true)"

# The scanner owner publishes the cause; Make's executor status is not a
# security verdict. The graph parent owns the canonical result and cleanup.
fail_with() {
  local failure_class="$1" failure_reason="$2" exit_code="$3"
  if [[ -n "$node_bin" ]] && ! "$node_bin" "$ROOT_DIR/tools/harness/runtime/command-failure-cli.mjs" "$failure_class" "$failure_reason"; then
    failure_class=harness
    failure_reason=scheduler_accounting_error
    exit_code=11
  fi
  printf 'failure_class=%s failure_reason=%s exit_code=%s\n' "$failure_class" "$failure_reason" "$exit_code" >&2
  exit "$exit_code"
}

if [[ -z "$node_bin" ]]; then
  echo "go-vulncheck requires node to parse Govulncheck JSON findings" >&2
  fail_with config configuration_error 2
fi

if [[ "$GO_BIN" != */* ]] && command -v "$GO_BIN" >/dev/null 2>&1; then
  GO_BIN="$(command -v "$GO_BIN")"
elif [[ "$GO_BIN" != /* ]]; then
  GO_BIN="$ROOT_DIR/$GO_BIN"
fi

if [[ ! -x "$GO_BIN" ]]; then
  echo "go-vulncheck requires an executable GO at $GO_BIN" >&2
  fail_with config configuration_error 2
fi

if [[ "$GOVULNCHECK_BIN" != */* ]] && command -v "$GOVULNCHECK_BIN" >/dev/null 2>&1; then
  GOVULNCHECK_BIN="$(command -v "$GOVULNCHECK_BIN")"
elif [[ "$GOVULNCHECK_BIN" != /* ]]; then
  GOVULNCHECK_BIN="$ROOT_DIR/$GOVULNCHECK_BIN"
fi

if [[ ! -x "$GOVULNCHECK_BIN" ]]; then
  echo "go-vulncheck requires an executable GOVULNCHECK_BIN at $GOVULNCHECK_BIN" >&2
  echo "run make go-security-toolchain before go-vulncheck or set GOVULNCHECK_BIN to a ready govulncheck binary" >&2
  fail_with config configuration_error 2
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
  fail_with config configuration_error 2
fi
read -r -a patterns <<<"$GOVULNCHECK_PATTERNS"

if ! package_output="$(
  GOCACHE="$GO_CACHE_DIR" \
  GOMODCACHE="$GO_MOD_CACHE_DIR" \
  GOTMPDIR="$GO_TMP_DIR" \
    "$GO_BIN" list "${patterns[@]}"
)"; then
  echo "go-vulncheck package discovery failed" >&2
  fail_with harness tool_diagnostic_failure 1
fi
mapfile -t packages < <(printf '%s\n' "$package_output" | cartulary_filter_authored_go_packages)

if [[ "${#packages[@]}" -eq 0 || -z "${packages[*]}" ]]; then
  echo "go-vulncheck package discovery returned no authored packages" >&2
  fail_with harness tool_diagnostic_failure 1
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

set +e
"$node_bin" "$ROOT_DIR/tools/harness/static-analysis/govulncheck-findings.mjs" \
  --input "$raw_output" \
  --output "$findings_output"
findings_status=$?
set -e

case "$findings_status" in
  0)
    if [[ "$scan_status" -ne 0 ]]; then
      fail_with harness tool_diagnostic_failure 1
    fi
    exit 0
    ;;
  1)
    fail_with security security_finding 1
    ;;
  *)
    if [[ "$scan_status" -ne 0 ]]; then
      fail_with harness tool_diagnostic_failure 1
    fi
    fail_with artifact artifact_error 11
    ;;
esac
