#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"

schema_id=""
scope=""
profile_id=""
cache_dir=""
disable_env=""
force_env=""
inputs=()
input_dirs=()
outputs=()
output_dirs=()
key_values=()
command=()

fail() {
  echo "$*" >&2
  exit 2
}

sha256_file() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
    return 0
  fi
  fail "sha256sum or shasum is required for cache hashing"
}

sha256_text_file() {
  local file="$1"
  sha256_file "$file"
}

repo_rel() {
  local path_value="$1"
  case "$path_value" in
    "$ROOT_DIR"/*) printf '%s\n' "${path_value#"$ROOT_DIR"/}" ;;
    *) printf '%s\n' "$path_value" ;;
  esac
}

sanitize() {
  printf '%s\n' "$1" | sed -E 's/[^A-Za-z0-9._-]+/-/g'
}

json_escape() {
  printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g'
}

json_string_array() {
  local first=1
  local value
  printf '['
  for value in "$@"; do
    if [[ "$first" -eq 0 ]]; then
      printf ','
    fi
    first=0
    printf '"%s"' "$(json_escape "$value")"
  done
  printf ']'
}

append_file_digest() {
  local material="$1"
  local label="$2"
  local file="$3"
  local display
  display="$(repo_rel "$file")"
  if [[ -f "$file" ]]; then
    printf '%s\t%s\tsha256:%s\n' "$label" "$display" "$(sha256_file "$file")" >>"$material"
  else
    printf '%s\t%s\tmissing\n' "$label" "$display" >>"$material"
  fi
}

append_dir_digest() {
  local material="$1"
  local label="$2"
  local dir="$3"
  local display
  local digest_file
  display="$(repo_rel "$dir")"
  if [[ ! -d "$dir" ]]; then
    printf '%s\t%s\tmissing\n' "$label" "$display" >>"$material"
    return
  fi
  digest_file="$(mktemp)"
  while IFS= read -r -d '' file; do
    printf 'file\t%s\tsha256:%s\n' "$(repo_rel "$file")" "$(sha256_file "$file")" >>"$digest_file"
  done < <(find "$dir" -type f -print0 | LC_ALL=C sort -z)
  printf '%s\t%s\tsha256:%s\n' "$label" "$display" "$(sha256_text_file "$digest_file")" >>"$material"
  rm -f "$digest_file"
}

output_digest() {
  local material
  local file
  local dir
  material="$(mktemp)"
  for file in "${outputs[@]}"; do
    append_file_digest "$material" "output" "$file"
  done
  for dir in "${output_dirs[@]}"; do
    append_dir_digest "$material" "output-dir" "$dir"
  done
  printf 'sha256:%s\n' "$(sha256_text_file "$material")"
  rm -f "$material"
}

outputs_missing() {
  local file
  local dir
  for file in "${outputs[@]}"; do
    [[ -f "$file" ]] || return 0
  done
  for dir in "${output_dirs[@]}"; do
    [[ -d "$dir" ]] || return 0
  done
  return 1
}

write_cache_json() {
  local file="$1"
  local state="$2"
  local reason="$3"
  local output_digest_value="$4"
  local created_at
  local output_entries=()
  local non_cacheable=(
    "public_target_summary"
    "failure_classification"
    "drift_security_service_cleanup_verdicts"
  )
  local entry
  created_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  for entry in "${outputs[@]}"; do
    output_entries+=("$(repo_rel "$entry")")
  done
  for entry in "${output_dirs[@]}"; do
    output_entries+=("$(repo_rel "$entry")")
  done
  mkdir -p "$(dirname "$file")"
  {
    printf '{\n'
    printf '  "schema_id": "%s",\n' "$(json_escape "$schema_id")"
    printf '  "cache_scope": "%s",\n' "$(json_escape "$scope")"
    printf '  "profile_id": "%s",\n' "$(json_escape "$profile_id")"
    printf '  "state": "%s",\n' "$(json_escape "$state")"
    printf '  "reason_code": "%s",\n' "$(json_escape "$reason")"
    printf '  "key_sha256": "sha256:%s",\n' "$key_hash"
    printf '  "input_digest_sha256": "sha256:%s",\n' "$input_hash"
    printf '  "output_digest_sha256": "%s",\n' "$(json_escape "$output_digest_value")"
    printf '  "record_path": "%s",\n' "$(json_escape "$(repo_rel "$record_file")")"
    printf '  "cacheable_outputs": '
    json_string_array "${output_entries[@]}"
    printf ',\n'
    printf '  "non_cacheable_side_effects": '
    json_string_array "${non_cacheable[@]}"
    printf ',\n'
    printf '  "created_at": "%s"\n' "$created_at"
    printf '}\n'
  } >"$file"
}

write_run_artifact() {
  local state="$1"
  local reason="$2"
  local output_digest_value="$3"
  local target="${CARTULARY_TEST_TARGET:-}"
  local run_root=""
  local artifact_file
  if [[ -z "${CARTULARY_TEST_RESULTS_DIR:-}" || -z "${CARTULARY_TEST_RUN_ID:-}" || -z "$target" ]]; then
    return
  fi
  run_root="${CARTULARY_TEST_RESULTS_DIR%/}/${CARTULARY_TEST_RUN_ID}"
  artifact_file="$run_root/$target/$(sanitize "$scope")-cache-$(sanitize "$profile_id").json"
  write_cache_json "$artifact_file" "$state" "$reason" "$output_digest_value"
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --schema-id) schema_id="${2:-}"; shift 2 ;;
    --scope) scope="${2:-}"; shift 2 ;;
    --profile) profile_id="${2:-}"; shift 2 ;;
    --cache-dir) cache_dir="${2:-}"; shift 2 ;;
    --disable-env) disable_env="${2:-}"; shift 2 ;;
    --force-env) force_env="${2:-}"; shift 2 ;;
    --input) inputs+=("${2:-}"); shift 2 ;;
    --input-dir) input_dirs+=("${2:-}"); shift 2 ;;
    --output) outputs+=("${2:-}"); shift 2 ;;
    --output-dir) output_dirs+=("${2:-}"); shift 2 ;;
    --key) key_values+=("${2:-}"); shift 2 ;;
    --) shift; command=("$@"); break ;;
    *) fail "unknown cache-artifact argument: $1" ;;
  esac
done

[[ -n "$schema_id" ]] || fail "--schema-id is required"
[[ -n "$scope" ]] || fail "--scope is required"
[[ -n "$profile_id" ]] || fail "--profile is required"
[[ "${#outputs[@]}" -gt 0 || "${#output_dirs[@]}" -gt 0 ]] || fail "at least one --output or --output-dir is required"
if [[ -z "$cache_dir" ]]; then
  cache_dir="$ROOT_DIR/.cache/cartulary/$scope"
fi

key_material="$(mktemp)"
printf 'schema_id\t%s\nscope\t%s\nprofile\t%s\nplatform\t%s:%s\n' \
  "$schema_id" "$scope" "$profile_id" "$(uname -s)" "$(uname -m)" >"$key_material"
printf 'helper\tsha256:%s\n' "$(sha256_file "$0")" >>"$key_material"
for key in "${key_values[@]}"; do
  printf 'key\t%s\n' "$key" >>"$key_material"
done
for input in "${inputs[@]}"; do
  append_file_digest "$key_material" "input" "$input"
done
for input_dir in "${input_dirs[@]}"; do
  append_dir_digest "$key_material" "input-dir" "$input_dir"
done
printf 'command' >>"$key_material"
for arg in "${command[@]}"; do
  printf '\t%s' "$arg" >>"$key_material"
done
printf '\n' >>"$key_material"

input_hash="$(sha256_text_file "$key_material")"
key_hash="$input_hash"
record_file="$cache_dir/$(sanitize "$profile_id")/$key_hash.json"

disabled=0
forced=0
if [[ -n "$disable_env" && "${!disable_env:-0}" == "1" ]]; then
  disabled=1
fi
if [[ -n "$force_env" && "${!force_env:-0}" == "1" ]]; then
  forced=1
fi

current_output_digest="$(output_digest)"
if [[ "$disabled" -eq 0 && "$forced" -eq 0 && -f "$record_file" ]] && ! outputs_missing; then
  if grep -Fq "\"schema_id\": \"$schema_id\"" "$record_file" &&
    grep -Fq "\"profile_id\": \"$profile_id\"" "$record_file" &&
    grep -Fq "\"key_sha256\": \"sha256:$key_hash\"" "$record_file" &&
    grep -Fq "\"output_digest_sha256\": \"$current_output_digest\"" "$record_file"; then
    write_run_artifact "hit" "cache_hit" "$current_output_digest"
    rm -f "$key_material"
    exit 0
  fi
  reason="cache_record_invalid"
else
  reason="cache_record_missing"
fi

if [[ "$disabled" -eq 1 ]]; then
  reason="env_disabled"
elif [[ "$forced" -eq 1 ]]; then
  reason="force_rebuild"
elif outputs_missing; then
  reason="output_missing"
elif [[ -f "$record_file" ]]; then
  reason="cache_record_invalid"
fi

if [[ "${#command[@]}" -eq 0 ]]; then
  fail "cache miss for $profile_id but no command was provided"
fi

"${command[@]}"

if outputs_missing; then
  rm -f "$key_material"
  fail "cache profile $profile_id command completed but declared outputs are missing"
fi

new_output_digest="$(output_digest)"
if [[ "$disabled" -eq 0 ]]; then
  write_cache_json "$record_file" "stored" "$reason" "$new_output_digest"
fi
write_run_artifact "$([[ "$disabled" -eq 1 ]] && printf 'disabled' || printf 'miss')" "$reason" "$new_output_digest"
rm -f "$key_material"
