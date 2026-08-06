#!/usr/bin/env bash
set -euo pipefail

mode="${1:-}"
case "$mode" in
  diagnose|ensure) ;;
  *)
    echo "usage: go-toolchain-readiness.sh diagnose|ensure" >&2
    exit 2
    ;;
esac

go_bin="${GO:-}"
expected_toolchain="${GO_TOOLCHAIN:-}"
go_cache_dir="${GO_CACHE_DIR:-}"
go_mod_cache_dir="${GO_MOD_CACHE_DIR:-}"

for required_name in GO GO_TOOLCHAIN GO_CACHE_DIR GO_MOD_CACHE_DIR; do
  if [[ -z "${!required_name:-}" ]]; then
    printf '%s is required for Go toolchain readiness\n' "$required_name" >&2
    exit 2
  fi
done

if [[ ! "$expected_toolchain" =~ ^go[1-9][0-9]*\.[0-9]+\.[0-9]+$ ]]; then
  echo "invalid pinned Go toolchain: ${expected_toolchain}" >&2
  exit 2
fi

if [[ ! -x "$go_bin" ]]; then
  if command -v "$go_bin" >/dev/null 2>&1; then
    go_bin="$(command -v "$go_bin")"
  else
    echo "missing go: install an automatic-toolchain-capable Go launcher or set GO=/path/to/go" >&2
    exit 2
  fi
fi

local_version_output=""
if ! local_version_output="$(env GOTOOLCHAIN=local GOTELEMETRY=off "$go_bin" version 2>&1)"; then
  printf 'Go launcher failed: %s\n' "$local_version_output" >&2
  exit 2
fi

read -r local_prefix local_word local_toolchain local_platform extra <<<"$local_version_output"
if [[ "$local_prefix" != "go" || "$local_word" != "version" || -z "$local_toolchain" || -z "$local_platform" || -n "${extra:-}" ]]; then
  printf 'Go launcher returned an invalid version line: %s\n' "$local_version_output" >&2
  exit 2
fi

platform_output=""
if ! platform_output="$(env GOTOOLCHAIN=local GOTELEMETRY=off "$go_bin" env GOOS GOARCH 2>&1)"; then
  printf 'Go launcher platform query failed: %s\n' "$platform_output" >&2
  exit 2
fi
mapfile -t platform_parts <<<"$platform_output"
if [[ "${#platform_parts[@]}" -ne 2 || -z "${platform_parts[0]}" || -z "${platform_parts[1]}" ]]; then
  printf 'Go launcher returned an invalid platform: %s\n' "$platform_output" >&2
  exit 2
fi
goos="${platform_parts[0]}"
goarch="${platform_parts[1]}"
expected_platform="${goos}/${goarch}"

module_version="v0.0.1-${expected_toolchain}.${goos}-${goarch}"
toolchain_dir="${go_mod_cache_dir}/golang.org/toolchain@${module_version}"
ziphash_file="${go_mod_cache_dir}/cache/download/golang.org/toolchain/@v/${module_version}.ziphash"
source_marker="${toolchain_dir}/src/_go.mod"
activation_marker="${toolchain_dir}/src/go.mod"
cached_go="${toolchain_dir}/bin/go"

print_corruption_diagnostic() {
  local reason="$1"
  printf 'corrupt Go automatic-toolchain cache: %s\n' "$reason" >&2
  printf '%s\n' 'Stop every Go and Make job, then move these exact entries to a temporary quarantine:' >&2
  printf '  %s\n' "$toolchain_dir" "$ziphash_file" >&2
  printf '%s\n' 'Do not edit go.mod, create _go.mod manually, disable checksum verification, or delete the broader module cache.' >&2
}

markers_match() {
  [[ -f "$source_marker" && -f "$activation_marker" ]] || return 1
  cmp -s -- "$source_marker" "$activation_marker"
}

if [[ "$local_toolchain" == "$expected_toolchain" ]]; then
  printf 'ok go: launcher=%s effective=%s source=local\n' "$local_toolchain" "$expected_toolchain"
  exit 0
fi

if [[ -d "$toolchain_dir" ]]; then
  if [[ ! -f "$source_marker" ]]; then
    print_corruption_diagnostic "missing ${source_marker}"
    exit 2
  fi
  if [[ ! -f "$ziphash_file" || ! -x "$cached_go" ]] || ! markers_match; then
    print_corruption_diagnostic "incomplete or inconsistent ${toolchain_dir}"
    exit 2
  fi
fi

if [[ "$mode" == "diagnose" ]]; then
  if [[ ! -d "$toolchain_dir" ]]; then
    printf 'missing go: launcher=%s; pinned effective toolchain %s is not installed in %s; run make bootstrap\n' \
      "$local_toolchain" "$expected_toolchain" "$go_mod_cache_dir" >&2
    exit 2
  fi
  cached_version_output=""
  if ! cached_version_output="$(env GOTOOLCHAIN=local GOTELEMETRY=off "$cached_go" version 2>&1)"; then
    printf 'cached Go toolchain failed: %s\n' "$cached_version_output" >&2
    exit 2
  fi
  read -r cached_prefix cached_word cached_toolchain cached_platform cached_extra <<<"$cached_version_output"
  if [[ "$cached_prefix" != "go" || "$cached_word" != "version" || "$cached_toolchain" != "$expected_toolchain" || "$cached_platform" != "$expected_platform" || -n "${cached_extra:-}" ]]; then
    printf 'cached Go toolchain mismatch: expected %s %s, got %s\n' \
      "$expected_toolchain" "$expected_platform" "$cached_version_output" >&2
    exit 2
  fi

  printf 'ok go: launcher=%s effective=%s source=automatic-cache\n' "$local_toolchain" "$expected_toolchain"
  exit 0
fi

if ! mkdir -p "$go_cache_dir" "$go_mod_cache_dir"; then
  printf 'cannot prepare Go cache paths %s and %s\n' "$go_cache_dir" "$go_mod_cache_dir" >&2
  exit 2
fi
effective_output=""
if ! effective_output="$(env GOTOOLCHAIN="$expected_toolchain" GOTELEMETRY=off GOCACHE="$go_cache_dir" GOMODCACHE="$go_mod_cache_dir" "$go_bin" version 2>&1)"; then
  if [[ -d "$toolchain_dir" && ! -f "$source_marker" ]]; then
    print_corruption_diagnostic "missing ${source_marker}"
  else
    printf 'Go toolchain readiness failed for %s: %s\n' "$expected_toolchain" "$effective_output" >&2
  fi
  exit 2
fi

effective_version_line="$(printf '%s\n' "$effective_output" | awk '$1 == "go" && $2 == "version" { line = $0 } END { print line }')"
read -r effective_prefix effective_word effective_toolchain effective_platform effective_extra <<<"$effective_version_line"
if [[ "$effective_prefix" != "go" || "$effective_word" != "version" || "$effective_toolchain" != "$expected_toolchain" || "$effective_platform" != "$expected_platform" || -n "${effective_extra:-}" ]]; then
  printf 'effective Go toolchain mismatch: expected %s %s, got %s\n' \
    "$expected_toolchain" "$expected_platform" "${effective_version_line:-$effective_output}" >&2
  exit 2
fi

printf 'ok go: launcher=%s effective=%s source=selected\n' "$local_toolchain" "$effective_toolchain"
