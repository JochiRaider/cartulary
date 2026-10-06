#!/usr/bin/env bash
# Test composition consumes a single Make producer receipt; it never builds images.
consume_package_release() {
  local root="$1" installation="$2" evidence="$3" extracted status=0
  extracted="$(mktemp -d "${TMPDIR:-/tmp}/cartulary-package-extract.XXXXXXXX")"
  if ! "${NODE_BIN:-node}" "$root/tools/release-evidence/package-release.mjs" unpack "$extracted" >"$evidence/package.json"; then status=1; fi
  if [[ "$status" == 0 ]]; then
    rmdir "$installation" || status=1
    if [[ "$status" == 0 ]] && ! "$extracted/release/assets/scripts/install.sh" "$installation" >"$evidence/installation.log"; then status=1; fi
  fi
  rm -rf -- "$extracted"
  [[ "$status" == 0 ]] || return 1
  # These caller bindings select only the installed, verified payload.
  # shellcheck disable=SC2034
  PACKAGE_DIR="$installation/release/assets"
  # shellcheck disable=SC2034
  package_image="$(jq -r '.images[] | select(.role=="application") | .id' "$evidence/package.json")"
  # shellcheck disable=SC2034
  package_postgres_image="$(jq -r '.images[] | select(.role=="postgres") | .id' "$evidence/package.json")"
  export CARTULARY_INSTALLATION_DIR="$installation"
}
