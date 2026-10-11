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
output_dir="$TMP_DIR/output-dir"
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
mkdir -p "$OUTPUT_DIR/empty"
cp "$INPUT_FILE" "$OUTPUT_DIR/copied.txt"
SH
chmod +x "$command_script"
printf 'value-one\n' >"$input_file"

run_cache() {
  CARTULARY_TEST_RESULTS_DIR="$results_dir" \
  CARTULARY_TEST_RUN_ID="$run_id" \
  CARTULARY_TEST_TARGET="$target" \
  INPUT_FILE="$input_file" \
  OUTPUT_FILE="$output_file" \
  OUTPUT_DIR="$output_dir" \
  COMMAND_LOG="$command_log" \
    "$SCRIPT" \
      \
      --scope "${fixture_scope:-build-artifact}" \
      --profile fixture-build \
      --cache-dir "$cache_dir" \
      --disable-env CARTULARY_BUILD_CACHE_DISABLE \
      --force-env CARTULARY_FORCE_REBUILD \
      --input "$input_file" \
      --input "$command_script" \
      --output "$output_file" \
      --output-dir "$output_dir" \
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
assert_equals "$(json_field "$record_path" 'value.artifacts[0].members.length')" "0" "file has no directory members"
assert_equals "$(json_field "$record_path" 'value.artifacts[1].members.map(member => member.relative_path).join(",")')" \
  "copied.txt,empty" "directory records both files and empty directories"

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

fixture_scope=readiness
set +e
CARTULARY_READINESS_CHECK_ONLY=1 run_cache >"$TMP_DIR/readiness.log" 2>&1
readiness_status=$?
set -e
assert_equals "$readiness_status" "2" "inspect-only readiness rejects a cold cache"
assert_equals "$(grep -c '^run$' "$command_log")" "6" "inspect-only readiness never invokes installation"
run_cache
CARTULARY_READINESS_CHECK_ONLY=1 run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "7" "inspect-only readiness accepts the exact warm cache"
unset fixture_scope

# Independent digest oracle and hash-process bound for a multi-batch input set.
# NUL records must preserve spaces, backslashes, newlines and option-like names.
batch_root="$TMP_DIR/batch"
mkdir -p "$batch_root/bin" "$batch_root/dir/empty"
real_sha256="$(command -v sha256sum)"
export REAL_SHA256="$real_sha256" HASH_CALLS="$batch_root/hash-calls"
cat >"$batch_root/bin/sha256sum" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" == "--zero" ]]; then
  printf '%s\n' "$(($# - 2))" >>"$HASH_CALLS"
  if [[ "${FAIL_HASH_BATCH:-0}" == "1" ]]; then exit 1; fi
fi
exec "$REAL_SHA256" "$@"
SH
chmod +x "$batch_root/bin/sha256sum"
batch_inputs=()
for ((i=0;i<260;i++)); do
  file="$batch_root/input-$i"
  printf 'content %s\n' "$i" >"$file"
  batch_inputs+=(--input "$file")
done
for name in 'space name' 'back\slash' $'line\nbreak' '-option'; do
  file="$batch_root/$name"
  printf 'special\n' >"$file"
  batch_inputs+=(--input "$file")
done
batch_inputs+=(--input "$batch_root/missing")
printf 'directory member\n' >"$batch_root/dir/member"
printf 'initial output\n' >"$batch_root/output"
cat >"$batch_root/build" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf 'run\n' >>"$BATCH_LOG"
printf 'output\n' >"$BATCH_OUTPUT"
SH
chmod +x "$batch_root/build"
export BATCH_LOG="$batch_root/build.log" BATCH_OUTPUT="$batch_root/output"
run_batch() {
  PATH="$batch_root/bin:$PATH" "$SCRIPT" --scope build-artifact --profile batch \
    --cache-dir "$batch_root/cache" "${batch_inputs[@]}" --input-dir "$batch_root/dir" \
    --output "$batch_root/output" --key fixture=batch -- "$batch_root/build"
}
run_batch
"$NODE_BIN" --input-type=module - "$ROOT_DIR" "$SCRIPT" "$batch_root" <<'JS'
import assert from "node:assert/strict";
import { readFileSync, readdirSync } from "node:fs";
import { createHash } from "node:crypto";
import os from "node:os";
const [root, script, dir] = process.argv.slice(2);
const hash = bytes => createHash("sha256").update(bytes).digest("hex");
const relative = file => file.startsWith(`${root}/`) ? file.slice(root.length + 1) : file;
const files = [...Array.from({length:260},(_,i)=>`${dir}/input-${i}`), ...["space name","back\\slash","line\nbreak","-option"].map(name=>`${dir}/${name}`)];
let material = `schema_id\tcartulary.harness_cache_record.v2\nscope\tbuild-artifact\nprofile\tbatch\nplatform\t${os.type()}:${os.machine()}\nhelper\tsha256:${hash(readFileSync(script))}\nkey\tfixture=batch\n`;
for (const file of files) material += `input\t${relative(file)}\tsha256:${hash(readFileSync(file))}\n`;
material += `input\t${relative(`${dir}/missing`)}\tmissing\n`;
const member = `file\t${relative(`${dir}/dir/member`)}\tsha256:${hash(readFileSync(`${dir}/dir/member`))}\n`;
material += `input-dir\t${relative(`${dir}/dir`)}\tsha256:${hash(member)}\ncommand\t${dir}/build\n`;
const records = readdirSync(`${dir}/cache/batch`);
assert.equal(records.length,1);
const record = JSON.parse(readFileSync(`${dir}/cache/batch/${records[0]}`));
assert.equal(record.input_digest, `sha256:${hash(material)}`);
assert.equal(record.unit_digest, record.input_digest);
assert.equal(records[0],`${hash(material)}.json`);
const calls = readFileSync(`${dir}/hash-calls`,"utf8").trim().split("\n").map(Number);
assert.deepEqual(calls,[128,128,8,1,1,1]);
JS
run_batch
assert_equals "$(grep -c '^run$' "$BATCH_LOG")" "1" "batched warm reuse"
printf 'changed output\n' >"$batch_root/output"
run_batch
assert_equals "$(grep -c '^run$' "$BATCH_LOG")" "2" "changed output rejects reuse"
if FAIL_HASH_BATCH=1 run_batch >"$batch_root/failure.log" 2>&1; then
  fail "failed digest batch was accepted"
fi
assert_equals "$(grep -c '^run$' "$BATCH_LOG")" "2" "hash failure never runs producer"
grep -Fq 'cache hashing failed' "$batch_root/failure.log" || fail "hash failure lost its cause"

bash "$ROOT_DIR/tools/harness/backend/tests/test-build-go-artifact.sh"

echo "cache artifact tests passed"
