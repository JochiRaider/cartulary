#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-$ROOT_DIR/tmp/node-runtime}"
PNPM_BIN="${PNPM:-$NODE_RUNTIME_DIR/bin/pnpm}"
CACHE_ARTIFACT_SCRIPT="$ROOT_DIR/tools/harness/readiness/cache-artifact.sh"
CACHE_DIR="${CARTULARY_STATIC_ANALYSIS_CACHE_DIR:-$ROOT_DIR/.cache/cartulary/static-analysis}"
CACHE_STAMP="$CACHE_DIR/outputs/lint-markdown.ok"

if [[ ! -x "$PNPM_BIN" ]]; then
  echo "repo-local pnpm was not found at $PNPM_BIN; run make frontend-toolchain" >&2
  exit 127
fi

status=0
markdown_inputs=(
  "$ROOT_DIR/.markdownlint-cli2.jsonc"
  "$ROOT_DIR/package.json"
  "$ROOT_DIR/pnpm-lock.yaml"
  "$ROOT_DIR/tools/harness/static-analysis/markdownlint.sh"
  "$ROOT_DIR/tools/harness/static-analysis/markdownlint-runner.sh"
  "$ROOT_DIR/tools/harness/readiness/cache-artifact.sh"
)
while IFS= read -r -d '' rel; do
  markdown_inputs+=("$ROOT_DIR/$rel")
done < <(git -C "$ROOT_DIR" ls-files -z --cached --others --exclude-standard -- '*.md' '*.markdown')

cache_args=(
  --scope static-analysis
  --profile lint-markdown
  --cache-dir "$CACHE_DIR"
  --disable-env CARTULARY_STATIC_ANALYSIS_DISABLE_CACHE
  --force-env CARTULARY_STATIC_ANALYSIS_FORCE
)
for input in "${markdown_inputs[@]}"; do
  [[ -e "$input" ]] || continue
  cache_args+=(--input "$input")
done
cache_args+=(
  --output "$CACHE_STAMP"
  --key "node_runtime_dir=$NODE_RUNTIME_DIR"
  --key "pnpm_bin=$PNPM_BIN"
  --key "args=$*"
)

env \
  ROOT_DIR="$ROOT_DIR" \
  NODE_RUNTIME_DIR="$NODE_RUNTIME_DIR" \
  PNPM_BIN="$PNPM_BIN" \
  CARTULARY_STATIC_CACHE_STAMP="$CACHE_STAMP" \
  PATH="$NODE_RUNTIME_DIR/bin:$PATH" \
  COREPACK_HOME="${COREPACK_HOME:-$NODE_RUNTIME_DIR/corepack}" \
  "$CACHE_ARTIFACT_SCRIPT" "${cache_args[@]}" -- "$ROOT_DIR/tools/harness/static-analysis/markdownlint-runner.sh" "$@" || status=$?

exit "$status"
