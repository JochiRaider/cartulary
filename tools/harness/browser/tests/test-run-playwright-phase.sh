#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
HELPER="$ROOT_DIR/tools/harness/browser/run-playwright-phase.sh"
cleanup_paths=()

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
  done
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_empty() {
  local value="$1"
  local label="$2"

  if [[ -n "$value" ]]; then
    fail "$label: expected no output, got [$value]"
  fi
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/run-playwright-phase-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")
attachment_file="$tmp_dir/trace.zip"
touch "$attachment_file"
fake_playwright="$tmp_dir/fake-playwright.sh"
cat >"$fake_playwright" <<EOF
#!/usr/bin/env bash
set -euo pipefail

output_file="\${PLAYWRIGHT_JSON_OUTPUT_FILE:-}"
if [[ -z "\$output_file" ]]; then
  echo "missing PLAYWRIGHT_JSON_OUTPUT_FILE" >&2
  exit 2
fi
mkdir -p "\$(dirname "\$output_file")"

case "\${FAKE_PLAYWRIGHT_MODE:-success}" in
  success)
    cat >"\$output_file" <<'JSON'
{"suites":[{"specs":[{"title":"raw success","file":"phase4.spec.ts","tests":[{"results":[{"status":"passed","retry":0,"attachments":[],"errors":[]}]}]}],"suites":[]}],"errors":[]}
JSON
    ;;
  failure)
    cat >"\$output_file" <<'JSON'
{"suites":[{"specs":[{"title":"raw failure","file":"phase4.spec.ts","tests":[{"results":[{"status":"failed","retry":0,"attachments":[{"name":"trace","path":"ATTACHMENT_FILE"}],"errors":[{"message":"Error: raw failure stack"}],"error":{"message":"Error: raw failure stack"}}]}]}],"suites":[]}],"errors":[]}
JSON
    sed -i "s|ATTACHMENT_FILE|$attachment_file|g" "\$output_file"
    exit 1
    ;;
  *)
    echo "unsupported fake playwright mode \${FAKE_PLAYWRIGHT_MODE}" >&2
    exit 2
    ;;
esac
EOF
chmod +x "$fake_playwright"

success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  NODE_BIN="${NODE:-node}" \
    "$HELPER" "playwright raw success" -- "$fake_playwright"
)"
assert_empty "$success_output" "playwright raw success"

set +e
failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="${NODE:-node}" \
  FAKE_PLAYWRIGHT_MODE=failure \
    "$HELPER" "playwright raw failure" -- "$fake_playwright" \
    2>&1
)"
failure_status=$?
set -e

if [[ "$failure_status" -eq 0 ]]; then
  fail "playwright raw failure: expected non-zero exit status"
fi
assert_contains "$failure_output" "failure: playwright raw failure" "playwright raw failure label"
assert_contains "$failure_output" "runner=playwright" "playwright raw failure runner"
assert_contains "$failure_output" "symbol_or_title=raw failure" "playwright raw failure title"
assert_contains "$failure_output" "retry=0" "playwright raw failure retry"
assert_contains "$failure_output" "trace.zip" "playwright raw attachment path"
