#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/workspace_layout.generated.sh
source "$(dirname -- "${BASH_SOURCE[0]}")/../../workspace_layout.generated.sh"

# Advisory static error-declaration audit (TH-HARNESS-REQ-867). Findings never
# change the exit status; a failed fixture proof, build or report shape does.
umask 077

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
GO_TMP_DIR="${GO_TMP_DIR:?GO_TMP_DIR is required}"
NODE_BIN="${NODE_BIN:?NODE_BIN is required}"
report_dir="${CARTULARY_STEP_ARTIFACT_DIR:?CARTULARY_STEP_ARTIFACT_DIR is required}"

if [[ "$GO_BIN" != */* ]] && command -v "$GO_BIN" >/dev/null 2>&1; then
  GO_BIN="$(command -v "$GO_BIN")"
elif [[ "$GO_BIN" != /* ]]; then
  GO_BIN="$ROOT_DIR/$GO_BIN"
fi
if [[ ! -x "$GO_BIN" ]]; then
  echo "go-error-declaration-audit requires an executable GO at $GO_BIN" >&2
  exit 1
fi

# The module graph comes only from the bootstrapped cache; the audit never
# resolves dependencies over the network.
go_env=(env GOWORK=off GOFLAGS=-mod=readonly GOPROXY=off GOTELEMETRY=off
  GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR")
analysis_dir="$ROOT_DIR/tools/go-analysis"

mkdir -p "$report_dir"
report="$report_dir/error-declaration-audit.json"

# The fixture proof must report its seeded undeclared code before the server
# program is analyzed.
"${go_env[@]}" "$GO_BIN" -C "$analysis_dir" test -count=1 ./errordeclaudit
"${go_env[@]}" "$GO_BIN" -C "$analysis_dir" run ./errordeclaudit -dir "$ROOT_DIR" -out "$report"
"$NODE_BIN" "$ROOT_DIR/tools/harness/static-analysis/error-declaration-audit-report.mjs" "$report"
