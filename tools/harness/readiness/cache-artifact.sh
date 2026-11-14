#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"

schema_id="cartulary.harness_cache_record.v2"
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

# The supported Linux profile uses one hash process per bounded batch. Hashing
# declared content remains mandatory; process creation must not scale per file.
sha256_file() {
  local record
  record="$(sha256sum -- "$1")" || fail "cache hashing failed"
  record="${record#\\}"
  [[ "${record:0:64}" =~ ^[0-9a-f]{64}$ ]] || fail "invalid cache digest"
  printf '%s\n' "${record:0:64}"
}

append_file_digests() (
  local material="$1" label="$2"
  shift 2
  local file display record digest_file index
  local -a batch=() present=()
  digest_file="$(mktemp "$scratch_root/digests.XXXXXX")" || fail "cache temporary allocation failed"
  trap 'rm -f -- "$digest_file"' EXIT
  while [[ "$#" -gt 0 ]]; do
    batch=("${@:1:128}")
    shift "${#batch[@]}"
    present=()
    for file in "${batch[@]}"; do
      if [[ -f "$file" ]]; then present+=("$file"); fi
    done
    : >"$digest_file" || fail "cache material write failed"
    if [[ "${#present[@]}" -gt 0 ]]; then
      sha256sum --zero -- "${present[@]}" >"$digest_file" || fail "cache hashing failed"
    fi
    index=0
    exec 3<"$digest_file"
    for file in "${batch[@]}"; do
      display="${file#"$ROOT_DIR"/}"
      if [[ "${present[$index]:-}" == "$file" ]]; then
        IFS= read -r -d '' record <&3 || fail "incomplete cache digest batch"
        [[ "${record:64:2}" == '  ' && "${record:66}" == "$file" && "${record:0:64}" =~ ^[0-9a-f]{64}$ ]] || fail "invalid cache digest batch"
        printf '%s\t%s\tsha256:%s\n' "$label" "$display" "${record:0:64}" >>"$material" || fail "cache material write failed"
        index=$((index + 1))
      else
        printf '%s\t%s\tmissing\n' "$label" "$display" >>"$material" || fail "cache material write failed"
      fi
    done
    if IFS= read -r -d '' record <&3 || [[ -n "$record" ]]; then fail "extra cache digest output"; fi
    exec 3<&-
  done
)

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
  local value="$1" code octal character escaped
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  for ((code=1; code<32; code++)); do
    printf -v octal '%03o' "$code"
    printf -v character '%b' "\\$octal"
    printf -v escaped '\\u%04x' "$code"
    value="${value//"$character"/"$escaped"}"
  done
  printf '%s' "$value"
}

assert_artifact_path() {
  local candidate="$1" parent
  while :; do
    [[ ! -L "$candidate" ]] || fail "cache artifact path must not contain links"
    parent="$(dirname -- "$candidate")"
    [[ "$parent" != "$candidate" ]] || break
    candidate="$parent"
  done
}

append_file_digest() {
  append_file_digests "$1" "$2" "$3"
}

append_dir_digest() {
  local material="$1"
  local label="$2"
  local dir="$3"
  local display
  local digest_file digest
  display="${dir#"$ROOT_DIR"/}"
  if [[ ! -d "$dir" ]]; then
    printf '%s\t%s\tmissing\n' "$label" "$display" >>"$material" || fail "cache material write failed"
    return
  fi
  digest_file="$(mktemp "$scratch_root/digests.XXXXXX")" || fail "cache temporary allocation failed"
  local files_file
  local -a files=()
  files_file="$(mktemp "$scratch_root/files.XXXXXX")" || fail "cache temporary allocation failed"
  if ! find "$dir" -type f -print0 | LC_ALL=C sort -z >"$files_file"; then
    rm -f -- "$files_file" "$digest_file"
    fail "cache directory discovery failed"
  fi
  mapfile -d '' -t files <"$files_file" || fail "cache directory discovery failed"
  rm -f -- "$files_file"
  append_file_digests "$digest_file" "file" "${files[@]}" || return $?
  digest="$(sha256_text_file "$digest_file")" || fail "cache hashing failed"
  printf '%s\t%s\tsha256:%s\n' "$label" "$display" "$digest" >>"$material" || fail "cache material write failed"
  rm -f "$digest_file"
}

output_digest() {
  local material digest
  local file
  local dir
  material="$(mktemp "$scratch_root/material.XXXXXX")" || fail "cache temporary allocation failed"
  for file in "${outputs[@]}"; do
    assert_artifact_path "$file" || return $?
    append_file_digest "$material" "output" "$file" || return $?
  done
  for dir in "${output_dirs[@]}"; do
    assert_artifact_path "$dir" || return $?
    append_dir_digest "$material" "output-dir" "$dir" || return $?
  done
  digest="$(sha256_text_file "$material")" || fail "cache hashing failed"
  printf 'sha256:%s\n' "$digest"
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

artifact_digest() {
  local path_value="$1"
  local material digest
  if [[ -f "$path_value" ]]; then
    digest="$(sha256_file "$path_value")" || fail "cache hashing failed"
    printf 'sha256:%s\n' "$digest"
    return
  fi
  material="$(mktemp "$scratch_root/material.XXXXXX")" || fail "cache temporary allocation failed"
  append_dir_digest "$material" "artifact" "$path_value" || return $?
  digest="$(sha256_text_file "$material")" || fail "cache hashing failed"
  printf 'sha256:%s\n' "$digest"
  rm -f "$material"
}

artifact_type() {
  [[ ! -L "$1" ]] || fail "cache artifact must not be a link"
  if [[ -f "$1" ]]; then
    printf 'file\n'
  elif [[ -d "$1" ]]; then
    printf 'directory\n'
  else
    fail "cache artifact must be a file or directory"
  fi
}

artifact_mode() {
  local mode
  mode="$(stat -c '%a' "$1")" || fail "cache artifact metadata failed"
  printf '0%s\n' "$mode"
}

write_artifact_members_json() {
  local root="$1"
  local member members_file kind mode digest
  local first=1
  printf '['
  if [[ -d "$root" ]]; then
    members_file="$(mktemp "$scratch_root/members.XXXXXX")" || fail "cache temporary allocation failed"
    find "$root" -mindepth 1 -print0 | LC_ALL=C sort -z >"$members_file" || fail "cache directory discovery failed"
    while IFS= read -r -d '' member; do
      if [[ "$first" -eq 0 ]]; then printf ','; fi
      first=0
      kind="$(artifact_type "$member")" || fail "cache artifact must be a file or directory without links"
      mode="$(artifact_mode "$member")" || fail "cache artifact metadata failed"
      digest="$(artifact_digest "$member")" || fail "cache hashing failed"
      printf '{"artifact_type":"%s","relative_path":"%s","mode":"%s","digest":"%s"}' \
        "$kind" \
        "$(json_escape "${member#"$root"/}")" \
        "$mode" \
        "$digest"
    done <"$members_file"
    rm -f -- "$members_file"
  fi
  printf ']'
}

write_artifacts_json() {
  local first=1
  local entry kind mode digest
  printf '['
  for entry in "${outputs[@]}" "${output_dirs[@]}"; do
    if [[ "$first" -eq 0 ]]; then
      printf ','
    fi
    first=0
    kind="$(artifact_type "$entry")" || fail "cache artifact must be a file or directory without links"
    mode="$(artifact_mode "$entry")" || fail "cache artifact metadata failed"
    digest="$(artifact_digest "$entry")" || fail "cache hashing failed"
    printf '{"artifact_type":"%s","relative_path":"%s","destination_class":"repository_artifact","mode":"%s","digest":"%s","producer_identity":"tool-cache:%s:%s","semantic_input_digest":"sha256:%s","members":' \
      "$kind" \
      "$(json_escape "$(repo_rel "$entry")")" \
      "$mode" \
      "$digest" \
      "$(json_escape "$scope")" \
      "$(json_escape "$profile_id")" \
      "$input_hash"
    write_artifact_members_json "$entry"
    printf '}'
  done
  printf ']'
}

render_cache_json() {
  local output_digest_value="$1" created_at="$2"
  {
    printf '{\n'
    printf '  "schema_id": "%s",\n' "$(json_escape "$schema_id")"
    printf '  "profile_id": "%s",\n' "$(json_escape "$profile_id")"
    printf '  "policy": "content_addressed",\n'
    printf '  "producer_identity": "tool-cache:%s:%s",\n' "$(json_escape "$scope")" "$(json_escape "$profile_id")"
    printf '  "unit_id": "tool-cache:%s:%s",\n' "$(json_escape "$scope")" "$(json_escape "$profile_id")"
    printf '  "unit_digest": "sha256:%s",\n' "$key_hash"
    printf '  "input_digest": "sha256:%s",\n' "$input_hash"
    printf '  "semantic_result": {"status":"passed","rows":[]},\n'
    printf '  "output_digest": "%s",\n' "$(json_escape "$output_digest_value")"
    printf '  "artifacts": '
    write_artifacts_json
    printf ',\n'
    printf '  "created_at": "%s"\n' "$created_at"
    printf '}\n'
  }
}

write_cache_json() {
  local file="$1" created_at
  created_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  mkdir -p "$(dirname "$file")"
  record_tmp="$(mktemp "${file}.tmp.XXXXXX")" || fail "cache temporary allocation failed"
  render_cache_json "$2" "$created_at" >"$record_tmp"
  mv -f "$record_tmp" "$file"
  record_tmp=""
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
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

[[ -n "$scope" ]] || fail "--scope is required"
[[ -n "$profile_id" ]] || fail "--profile is required"
[[ "${#outputs[@]}" -gt 0 || "${#output_dirs[@]}" -gt 0 ]] || fail "at least one --output or --output-dir is required"
if [[ -z "$cache_dir" ]]; then
  cache_dir="$ROOT_DIR/.cache/cartulary/$scope"
fi

# A profile owns its output transaction even when source keys change or caching
# is bypassed. Lock its stable directory through inspection and publication;
# locking only inside the producer leaves its completed proof exposed to a peer.
profile_dir="$cache_dir/$(sanitize "$profile_id")"
mkdir -p -- "$profile_dir"
exec {profile_lock}<"$profile_dir"
flock -x "$profile_lock"

scratch_root="$(mktemp -d)" || fail "cache temporary allocation failed"
record_tmp=""
cleanup() {
  if [[ -n "$record_tmp" ]]; then rm -f -- "$record_tmp"; fi
  rm -rf -- "$scratch_root"
}
trap cleanup EXIT

key_material="$(mktemp "$scratch_root/key.XXXXXX")" || fail "cache temporary allocation failed"
printf 'schema_id\t%s\nscope\t%s\nprofile\t%s\nplatform\t%s:%s\n' \
  "$schema_id" "$scope" "$profile_id" "$(uname -s)" "$(uname -m)" >"$key_material"
helper_digest="$(sha256_file "$0")" || fail "cache hashing failed"
printf 'helper\tsha256:%s\n' "$helper_digest" >>"$key_material"
for key in "${key_values[@]}"; do
  printf 'key\t%s\n' "$key" >>"$key_material"
done
append_file_digests "$key_material" "input" "${inputs[@]}"
for input_dir in "${input_dirs[@]}"; do
  append_dir_digest "$key_material" "input-dir" "$input_dir"
done
for output in "${outputs[@]}"; do
  printf 'output-contract\tfile\t%s\n' "$(repo_rel "$output")" >>"$key_material"
done
for output in "${output_dirs[@]}"; do
  printf 'output-contract\tdirectory\t%s\n' "$(repo_rel "$output")" >>"$key_material"
done
printf 'command' >>"$key_material"
for arg in "${command[@]}"; do
  printf '\t%s' "$arg" >>"$key_material"
done
printf '\n' >>"$key_material"

input_hash="$(sha256_text_file "$key_material")"
key_hash="$input_hash"
record_file="$profile_dir/$key_hash.json"

disabled=0
forced=0
if [[ -n "$disable_env" && "${!disable_env:-0}" == "1" ]]; then
  disabled=1
fi
if [[ -n "$force_env" && "${!force_env:-0}" == "1" ]]; then
  forced=1
fi

current_output_digest="$(output_digest)" || fail "cache hashing failed"
if [[ "$disabled" -eq 0 && "$forced" -eq 0 && -f "$record_file" && ! -L "$record_file" ]] && ! outputs_missing; then
  # Records are private canonical producer output. Exact comparison closes all
  # fields, duplicate/unknown members and trailing bytes without a bootstrap parser.
  recorded_at="$(sed -n 's/^  "created_at": "\([0-9TZ:-]*\)"$/\1/p' "$record_file")"
  if [[ "$recorded_at" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]] &&
      canonical_time="$(date -u -d "$recorded_at" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null)" && [[ "$canonical_time" == "$recorded_at" ]]; then
    expected_record="$scratch_root/expected.json"
    render_cache_json "$current_output_digest" "$recorded_at" >"$expected_record"
    if cmp -s -- "$record_file" "$expected_record"; then
      exit 0
    fi
  fi
fi

if [[ "$scope" == "readiness" && "${CARTULARY_READINESS_CHECK_ONLY:-0}" == "1" ]]; then
  rm -f "$key_material"
  echo "readiness cache miss for $profile_id; run the public bootstrap or install target" >&2
  exit 2
fi

if [[ "${#command[@]}" -eq 0 ]]; then
  fail "cache miss for $profile_id but no command was provided"
fi

"${command[@]}"

if outputs_missing; then
  rm -f "$key_material"
  fail "cache profile $profile_id command completed but declared outputs are missing"
fi

new_output_digest="$(output_digest)" || fail "cache hashing failed"
if [[ "$disabled" -eq 0 ]]; then
  write_cache_json "$record_file" "$new_output_digest"
fi
rm -f "$key_material"
