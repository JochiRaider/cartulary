#!/usr/bin/env bash

fail() {
  echo "$*" >&2
  exit 1
}

text_contains() {
  local text="$1"
  local needle="$2"

  [[ "$text" == *"$needle"* ]]
}

text_matches() {
  local text="$1"
  local pattern="$2"

  [[ "$text" =~ $pattern ]]
}

assert_text_contains() {
  local label="$1"
  local text="$2"
  local needle="$3"
  local message="$4"

  if ! text_contains "$text" "$needle"; then
    fail "${message:-$label must contain $needle}"
  fi
}

assert_text_not_contains() {
  local label="$1"
  local text="$2"
  local needle="$3"
  local message="$4"

  if text_contains "$text" "$needle"; then
    fail "${message:-$label must not contain $needle}"
  fi
}

assert_text_matches() {
  local label="$1"
  local text="$2"
  local pattern="$3"
  local message="$4"

  if ! text_matches "$text" "$pattern"; then
    fail "${message:-$label must match $pattern}"
  fi
}

assert_text_not_matches() {
  local label="$1"
  local text="$2"
  local pattern="$3"
  local message="$4"

  if text_matches "$text" "$pattern"; then
    fail "${message:-$label must not match $pattern}"
  fi
}

text_has_token() {
  local text="$1"
  local expected="$2"

  awk -v token="$expected" '
    {
      for (i = 1; i <= NF; i++) {
        if ($i == token) {
          found = 1
        }
      }
    }
    END { exit found ? 0 : 1 }
  ' <<<"$text"
}

task_surface_check_files() {
  local files=()

  if [[ -n "${generated_make:-}" ]]; then
    files+=("$generated_make")
  fi
  if [[ -n "${makefile:-}" ]]; then
    files+=("$makefile")
  fi
  if [[ "${#files[@]}" -eq 0 ]]; then
    fail "task-surface check must configure generated_make and makefile"
  fi

  printf '%s\n' "${files[@]}"
}

extract_target_block() {
  local target="$1"
  local files=()
  mapfile -t files < <(task_surface_check_files)

  awk -v target="$target" '
    $0 ~ "^" target ":[[:space:]]+export[[:space:]]" { next }
    $0 ~ "^" target ":" { in_block=1; next }
    in_block && /^[^[:space:]].*:/ { exit }
    in_block { print }
  ' "${files[@]}"
}

extract_target_prereqs() {
  local target="$1"
  local files=()
  mapfile -t files < <(task_surface_check_files)

  awk -v target="$target" '
    $0 ~ "^" target ":" && $0 !~ "^" target ":[[:space:]]+export[[:space:]]" {
      sub("^" target ":[[:space:]]*", "", $0)
      print
      exit
    }
  ' "${files[@]}"
}

target_exists() {
  local target="$1"
  local files=()
  mapfile -t files < <(task_surface_check_files)

  awk -v target="$target" '
    $0 ~ "^" target ":" && $0 !~ "^" target ":[[:space:]]+export[[:space:]]" {
      found=1
      exit
    }
    END { exit found ? 0 : 1 }
  ' "${files[@]}"
}

assert_target_exists() {
  local target="$1"
  local message="${2:-Makefile must define $target}"

  if ! target_exists "$target"; then
    fail "$message"
  fi
}

assert_target_absent() {
  local target="$1"
  local message="${2:-Makefile must not define $target}"

  if target_exists "$target"; then
    fail "$message"
  fi
}

assert_target_exports_self() {
  local target="$1"
  local message="${2:-$target must export CARTULARY_TEST_TARGET}"
  local files=()
  mapfile -t files < <(task_surface_check_files)

  if ! awk -v target="$target" '
    $0 == target ": export CARTULARY_TEST_TARGET ?= " target {
      found=1
      exit
    }
    END { exit found ? 0 : 1 }
  ' "${files[@]}"; then
    fail "$message"
  fi
}

assert_text_has_token() {
  local label="$1"
  local text="$2"
  local token="$3"
  local message="$4"

  if ! text_has_token "$text" "$token"; then
    fail "${message} (${label} missing ${token})"
  fi
}

assert_text_lacks_token() {
  local label="$1"
  local text="$2"
  local token="$3"
  local message="$4"

  if text_has_token "$text" "$token"; then
    fail "${message} (${label} unexpectedly included ${token})"
  fi
}

assert_target_prereq() {
  local target="$1"
  local prereq="$2"
  local message="$3"
  local prereqs

  prereqs="$(extract_target_prereqs "$target")"
  if [[ -z "$prereqs" ]]; then
    fail "Makefile must define non-empty $target prerequisites"
  fi
  assert_text_has_token "$target prerequisites" "$prereqs" "$prereq" "$message"
}
