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
  scratch_root="$(__cartulary_harness_realpath "${CARTULARY_HARNESS_SCRATCH_ROOT:-${TMPDIR:-/tmp}/cartulary-harness-scratch}")" || return 2

  case "${scratch_root}" in
    "${repo_root}"|"${repo_root}"/*)
      echo "CARTULARY_HARNESS_SCRATCH_ROOT must be outside the repository: ${scratch_root}" >&2
      return 2
      ;;
  esac

  mkdir -p "${scratch_root}"
  mktemp -d "${scratch_root}/${template}"
}
