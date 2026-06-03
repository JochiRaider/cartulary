#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="$ROOT_DIR/scripts/cache-artifact.sh"
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
      --schema-id cartulary.cache.build_artifact.v1 \
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

artifact_file="$results_dir/$run_id/$target/build-artifact-cache-fixture-build.json"

run_cache
assert_equals "$(cat "$command_log")" "run" "first miss executes command once"
assert_equals "$(json_field "$artifact_file" 'value.state')" "miss" "first run records miss"
record_path="$ROOT_DIR/$(json_field "$artifact_file" 'value.record_path')"
"$ROOT_DIR/scripts/harness-contract.sh" validate-schema cartulary.cache.build_artifact.v1 "$record_path" >/dev/null

run_cache
assert_equals "$(cat "$command_log")" "run" "cache hit skips command"
assert_equals "$(json_field "$artifact_file" 'value.state')" "hit" "second run records hit"

printf 'value-two\n' >"$input_file"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "2" "input change executes command"
assert_equals "$(json_field "$artifact_file" 'value.state')" "miss" "input change records miss"
record_path="$ROOT_DIR/$(json_field "$artifact_file" 'value.record_path')"

printf '{not valid json\n' >"$record_path"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "3" "corrupt record executes command"
assert_equals "$(json_field "$artifact_file" 'value.reason_code')" "cache_record_invalid" "corrupt record reason"

CARTULARY_BUILD_CACHE_DISABLE=1 run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "4" "disabled cache executes command"
assert_equals "$(json_field "$artifact_file" 'value.state')" "disabled" "disabled cache state"

CARTULARY_FORCE_REBUILD=1 run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "5" "force rebuild executes command"
assert_equals "$(json_field "$artifact_file" 'value.reason_code')" "force_rebuild" "force rebuild reason"

rm -f "$output_file"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "6" "missing output executes command"
assert_equals "$(json_field "$artifact_file" 'value.reason_code')" "output_missing" "missing output reason"

echo "cache artifact tests passed"
