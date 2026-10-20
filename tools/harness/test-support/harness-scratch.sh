#!/usr/bin/env bash

__cartulary_harness_default_repo_root() {
  unset CDPATH && cd -- "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P
}

__cartulary_harness_realpath() {
  if ! command -v realpath >/dev/null 2>&1; then
    echo "realpath is required to validate harness scratch paths" >&2
    return 2
  fi
  realpath -m "$1"
}

cartulary_harness_mktemp_dir() {
  if [[ "$#" -ne 1 ]]; then
    echo "cartulary_harness_mktemp_dir requires <template>" >&2
    return 2
  fi

  local template="$1"
  if [[ "${template}" == /* || "${template}" == *"/"* || "${template}" != *XXXXXX* ]]; then
    echo "cartulary_harness_mktemp_dir template must be a basename containing XXXXXX" >&2
    return 2
  fi

  local repo_root
  repo_root="$(__cartulary_harness_realpath "${CARTULARY_HARNESS_REPO_ROOT:-$(__cartulary_harness_default_repo_root)}")" || return 2

  local scratch_root
  if [[ -v CARTULARY_HARNESS_SCRATCH_ROOT ]]; then
    scratch_root="${CARTULARY_HARNESS_SCRATCH_ROOT}"
  else
    local machine_cache
    if [[ -v CARTULARY_MACHINE_CACHE_DIR ]]; then
      machine_cache="${CARTULARY_MACHINE_CACHE_DIR}"
    elif [[ "${XDG_CACHE_HOME:-}" == /* ]]; then
      machine_cache="${XDG_CACHE_HOME}/cartulary"
    elif [[ "${HOME:-}" == /* ]]; then
      machine_cache="${HOME}/.cache/cartulary"
    else
      echo "harness scratch requires an absolute machine-cache or scratch root" >&2
      return 2
    fi
    [[ "${machine_cache}" == /* ]] || return 2
    scratch_root="${machine_cache}/harness-scratch"
  fi
  if [[ "${scratch_root}" != /* ]]; then
    echo "CARTULARY_HARNESS_SCRATCH_ROOT must be absolute and nonempty" >&2
    return 2
  fi
  scratch_root="$(__cartulary_harness_realpath "${scratch_root}")" || return 2

  case "${scratch_root}" in
    "${repo_root}"|"${repo_root}"/*)
      echo "CARTULARY_HARNESS_SCRATCH_ROOT must be outside the repository: ${scratch_root}" >&2
      return 2
      ;;
  esac

  mkdir -p "${scratch_root}"
  mktemp -d "${scratch_root}/${template}"
}
