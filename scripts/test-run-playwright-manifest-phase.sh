#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-playwright-manifest-phase.sh"
cleanup_paths=()

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

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/run-playwright-manifest-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")
fake_playwright="$tmp_dir/fake-playwright.sh"
cat >"$fake_playwright" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

output_file="${PLAYWRIGHT_JSON_OUTPUT_FILE:-}"
if [[ -z "$output_file" ]]; then
  echo "missing PLAYWRIGHT_JSON_OUTPUT_FILE" >&2
  exit 2
fi
mkdir -p "$(dirname "$output_file")"

if [[ " $* " == *" --list "* ]]; then
  cat >"$output_file" <<'JSON'
{"suites":[{"specs":[{"title":"E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface","tests":[{"results":[],"status":"skipped"}],"file":"phase2.spec.ts"},{"title":"E-2-02 shows incident discovery, direct retrieval, and promoted-field-only patching on the ordinary incident shell","tests":[{"results":[],"status":"skipped"}],"file":"phase2.spec.ts"},{"title":"E-2-03 lets incident admins manage memberships and hides those controls from non-admin members on the ordinary shell","tests":[{"results":[],"status":"skipped"}],"file":"phase2.spec.ts"}],"suites":[]}],"errors":[]}
JSON
  exit 0
fi

case "${FAKE_PLAYWRIGHT_MODE:-success}" in
  success)
    cat >"$output_file" <<'JSON'
{"suites":[{"specs":[{"title":"E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface","file":"phase2.spec.ts","tests":[{"results":[{"status":"passed","retry":0,"attachments":[],"errors":[]}]}]},{"title":"E-2-02 shows incident discovery, direct retrieval, and promoted-field-only patching on the ordinary incident shell","file":"phase2.spec.ts","tests":[{"results":[{"status":"passed","retry":0,"attachments":[],"errors":[]}]}]},{"title":"E-2-03 lets incident admins manage memberships and hides those controls from non-admin members on the ordinary shell","file":"phase2.spec.ts","tests":[{"results":[{"status":"passed","retry":0,"attachments":[],"errors":[]}]}]}],"suites":[]}],"errors":[]}
JSON
    ;;
  mismatch)
    cat >"$output_file" <<'JSON'
{"suites":[{"specs":[{"title":"E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface","file":"phase2.spec.ts","tests":[{"results":[{"status":"passed","retry":0,"attachments":[],"errors":[]}]}]}],"suites":[]}],"errors":[]}
JSON
    ;;
  *)
    echo "unsupported fake playwright mode ${FAKE_PLAYWRIGHT_MODE}" >&2
    exit 2
    ;;
esac
EOF
chmod +x "$fake_playwright"

success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="${NODE:-node}" \
    "$HELPER" "playwright manifest success" phase2 authoritative -- "$fake_playwright"
)"
assert_empty "$success_output" "playwright manifest success"

set +e
mismatch_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="${NODE:-node}" \
  FAKE_PLAYWRIGHT_MODE=mismatch \
    "$HELPER" "playwright manifest mismatch" phase2 authoritative -- "$fake_playwright" \
    2>&1
)"
mismatch_status=$?
set -e

if [[ "$mismatch_status" -eq 0 ]]; then
  fail "playwright manifest mismatch: expected non-zero exit status"
fi
assert_contains "$mismatch_output" "manifest mismatch: playwright manifest mismatch" "playwright manifest mismatch label"
assert_contains "$mismatch_output" "missing_ids=E-2-02,E-2-03" "playwright manifest missing ids"
