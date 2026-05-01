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
