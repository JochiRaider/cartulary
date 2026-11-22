#!/usr/bin/env bash

# Root package patterns stop at module boundaries. Run discovery and its
# consumer in the same module so go and Staticcheck load the same graph.
cartulary_lint_go_modules() (
  local root_dir="$1" filter="$2" go_bin="${GO:-go}" module package_output
  shift 2
  local -a scanner=("$@") patterns packages
  if [[ "$go_bin" == */* && "$go_bin" != /* ]]; then
    go_bin="$root_dir/$go_bin"
  fi
  if [[ "${scanner[0]}" == */* && "${scanner[0]}" != /* ]]; then
    scanner[0]="$root_dir/${scanner[0]}"
  fi
  export GOCACHE="${GO_CACHE_DIR:?}" GOMODCACHE="${GO_MOD_CACHE_DIR:?}" GOTMPDIR="${GO_TMP_DIR:?}"

  for module in . tools/go-analysis; do
    cd -- "$root_dir/$module" || return 1
    if [[ "$module" == . ]]; then
      patterns=(./cmd/... ./internal/... ./db/... ./tools/...)
    else
      patterns=(./...)
      export GOWORK=off GOPROXY=off GOFLAGS=-mod=readonly GOTELEMETRY=off
    fi
    if ! package_output="$("$go_bin" list "${patterns[@]}" | "$filter")"; then
      echo "Go lint package discovery failed in $module" >&2
      return 1
    fi
    if [[ -z "$package_output" ]]; then
      echo "Go lint package discovery returned no authored packages in $module" >&2
      return 1
    fi
    mapfile -t packages <<< "$package_output"
    "${scanner[@]}" "${packages[@]}" || return "$?"
  done
)
