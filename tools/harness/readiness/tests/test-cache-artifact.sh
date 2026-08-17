#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="$ROOT_DIR/tools/harness/readiness/cache-artifact.sh"
NODE_BIN="${NODE_BIN:-node}"
TMP_DIR="$(mktemp -d "$ROOT_DIR/tmp/cache-artifact-test.XXXXXX")"

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    fail "$label: expected [$expected], got [$actual]"
  fi
}

json_field() {
  local file="$1"
  local expr="$2"
  "$NODE_BIN" - "$file" "$expr" <<'JS'
const fs = require("node:fs");
const [file, expr] = process.argv.slice(2);
const value = JSON.parse(fs.readFileSync(file, "utf8"));
const result = Function("value", `return ${expr}`)(value);
if (result === null || result === undefined) {
  process.stdout.write("");
} else {
  process.stdout.write(String(result));
}
JS
}

cache_dir="$TMP_DIR/cache"
input_file="$TMP_DIR/input.txt"
output_file="$TMP_DIR/output.txt"
command_log="$TMP_DIR/command.log"
command_script="$TMP_DIR/write-output.sh"
results_dir="$TMP_DIR/results"
run_id="run"
target="cache-artifact-fixture"

cat >"$command_script" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf 'run\n' >>"$COMMAND_LOG"
cp "$INPUT_FILE" "$OUTPUT_FILE"
SH
chmod +x "$command_script"
printf 'value-one\n' >"$input_file"

run_cache() {
  CARTULARY_TEST_RESULTS_DIR="$results_dir" \
  CARTULARY_TEST_RUN_ID="$run_id" \
  CARTULARY_TEST_TARGET="$target" \
  INPUT_FILE="$input_file" \
  OUTPUT_FILE="$output_file" \
  COMMAND_LOG="$command_log" \
    "$SCRIPT" \
      \
      --scope build-artifact \
      --profile fixture-build \
      --cache-dir "$cache_dir" \
      --disable-env CARTULARY_BUILD_CACHE_DISABLE \
      --force-env CARTULARY_FORCE_REBUILD \
      --input "$input_file" \
      --input "$command_script" \
      --output "$output_file" \
      --key "fixture=one" \
      -- "$command_script"
}

run_cache
assert_equals "$(cat "$command_log")" "run" "first miss executes command once"
record_path="$(find "$cache_dir/fixture-build" -type f -name '*.json' -print -quit)"
[[ -n "$record_path" ]] || fail "first miss did not publish a cache record"
"$NODE_BIN" "$ROOT_DIR/tools/harness/contract/harness-contract-cli.mjs" \
  validate-schema cartulary.harness_cache_record.v2 "$record_path" >/dev/null
assert_equals \
  "$(json_field "$record_path" 'value.policy')" \
  "content_addressed" \
  "record uses the canonical cache policy"
assert_equals \
  "$(json_field "$record_path" 'value.artifacts[0].relative_path')" \
  "${output_file#"$ROOT_DIR"/}" \
  "record identifies the cached output"

run_cache
assert_equals "$(cat "$command_log")" "run" "cache hit skips command"

printf 'value-two\n' >"$input_file"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "2" "input change executes command"
record_path="$(find "$cache_dir/fixture-build" -type f -name '*.json' -printf '%T@ %p\n' | sort -nr | head -1 | cut -d' ' -f2-)"

printf '{not valid json\n' >"$record_path"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "3" "corrupt record executes command"
"$NODE_BIN" "$ROOT_DIR/tools/harness/contract/harness-contract-cli.mjs" \
  validate-schema cartulary.harness_cache_record.v2 "$record_path" >/dev/null

CARTULARY_BUILD_CACHE_DISABLE=1 run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "4" "disabled cache executes command"

CARTULARY_FORCE_REBUILD=1 run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "5" "force rebuild executes command"

rm -f "$output_file"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "6" "missing output executes command"

bash "$ROOT_DIR/tools/harness/backend/tests/test-build-go-artifact.sh"

echo "cache artifact tests passed"
