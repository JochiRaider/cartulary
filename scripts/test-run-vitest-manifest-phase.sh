#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-vitest-manifest-phase.sh"
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

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/run-vitest-manifest-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")
fake_vitest="$tmp_dir/fake-vitest.sh"
cat >"$fake_vitest" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

cat <<'JSON'
{"numTotalTestSuites":2,"numPassedTestSuites":2,"numFailedTestSuites":0,"numPendingTestSuites":0,"numTotalTests":4,"numPassedTests":1,"numFailedTests":0,"numPendingTests":3,"numTodoTests":0,"success":true,"testResults":[{"assertionResults":[{"ancestorTitles":["Phase 3 Timeline workbook"],"fullName":"Phase 3 Timeline workbook Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels","status":"passed","title":"Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels","failureMessages":[],"meta":{},"tags":[]},{"ancestorTitles":["Phase 3 Timeline workbook"],"fullName":"Phase 3 Timeline workbook Phase 3 support keeps a continuation row helper after autosaved create","status":"skipped","title":"Phase 3 support keeps a continuation row helper after autosaved create","failureMessages":[],"meta":{},"tags":[]}],"status":"passed","message":"","name":"/home/askahn/code/cartulary/apps/web/src/App.test.tsx"}]}
JSON
EOF
chmod +x "$fake_vitest"

output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="${NODE:-node}" \
    "$HELPER" "vitest manifest smoke" phase3 authoritative frontend_unit -- "$fake_vitest"
)"

assert_contains "$output" "== vitest manifest smoke ==" "vitest manifest banner"
assert_contains "$output" "matched vitest manifest tests: 1" "vitest manifest matched count"
assert_contains "$output" "Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels" "vitest manifest title"
