#!/usr/bin/env bash
# Sourced only inside a Make-owned, private disposable package workspace.
provision_package_fixture_tls() {
  local root="$1" work="$3"
  (
    cd "$root" || return
    GOCACHE="${GO_CACHE_DIR:?}" GOMODCACHE="${GO_MOD_CACHE_DIR:?}" GOTMPDIR="${GO_TMP_DIR:?}" \
      "${GO:-go}" run ./tools/packagepki --directory "$work/tls"
  )
}
