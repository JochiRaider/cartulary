#!/usr/bin/env bash
# Sourced only inside a Make-owned, private disposable package workspace.
provision_package_fixture_tls() {
  local root="$1" package="$2" work="$3" name
  for name in postgres-provision.sh postgres-provision.sql postgres-entrypoint.sh postgres-hba.conf postgres-ident.conf seaweed-entrypoint.sh; do
    cp "$package/$name" "$work/$name"
    chmod 0644 "$work/$name"
  done
  chmod 0755 "$work/postgres-provision.sh" "$work/postgres-entrypoint.sh" "$work/seaweed-entrypoint.sh"
  (
    cd "$root" || return
    GOCACHE="${GO_CACHE_DIR:?}" GOMODCACHE="${GO_MOD_CACHE_DIR:?}" GOTMPDIR="${GO_TMP_DIR:?}" \
      "${GO:-go}" run ./tools/packagepki --directory "$work/tls"
  )
}
