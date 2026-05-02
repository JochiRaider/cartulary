#!/usr/bin/env bash

cartulary_is_generated_go_path() {
  local path="$1"
  case "$path" in
    internal/gen/*) return 0 ;;
    *) return 1 ;;
  esac
}

cartulary_has_generated_go_marker() {
  local path="$1"
  sed -n '1,20p' -- "$path" | grep -Eq '^// Code generated .* DO NOT EDIT\.$'
}

cartulary_is_authored_go_file() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    return 1
  fi
  if cartulary_is_generated_go_path "$path"; then
    return 1
  fi
  if cartulary_has_generated_go_marker "$path"; then
    return 1
  fi
  return 0
}

cartulary_is_generated_go_package() {
  local package="$1"
  case "$package" in
    */internal/gen | */internal/gen/*) return 0 ;;
    *) return 1 ;;
  esac
}

cartulary_filter_authored_go_packages() {
  local package
  while IFS= read -r package; do
    if cartulary_is_generated_go_package "$package"; then
      continue
    fi
    printf '%s\n' "$package"
  done
}

cartulary_is_generated_artifact_path() {
  local path="$1"
  case "$path" in
    internal/gen/* | packages/protocol-ts/src/generated/*) return 0 ;;
    *) return 1 ;;
  esac
}
