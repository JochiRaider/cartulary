#!/usr/bin/env bash
set -euo pipefail

source_index="${WEB_DIST_INDEX:?WEB_DIST_INDEX is required}"
asset_dir="${EMBEDDED_WEB_ASSET_DIR:?EMBEDDED_WEB_ASSET_DIR is required}"
asset_archive="${EMBEDDED_WEB_ASSET_ARCHIVE:?EMBEDDED_WEB_ASSET_ARCHIVE is required}"
asset_manifest="${EMBEDDED_CLIENT_ASSET_MANIFEST:?EMBEDDED_CLIENT_ASSET_MANIFEST is required}"
client_support_registry="${EMBEDDED_CLIENT_SUPPORT_REGISTRY:?EMBEDDED_CLIENT_SUPPORT_REGISTRY is required}"
client_support_source="${EXTENSION_CLIENT_SUPPORT_SOURCE:?EXTENSION_CLIENT_SUPPORT_SOURCE is required}"
asset_stamp="${EMBEDDED_WEB_ASSET_STAMP:?EMBEDDED_WEB_ASSET_STAMP is required}"
asset_ready_stamp="${EMBEDDED_WEB_ASSET_READY_STAMP:?EMBEDDED_WEB_ASSET_READY_STAMP is required}"
go_bin="${GO:?GO is required}"
go_cache_dir="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
go_mod_cache_dir="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
go_tmp_dir="${GO_TMP_DIR:?GO_TMP_DIR is required}"

mkdir -p "$asset_dir" "$(dirname "$asset_stamp")" "$(dirname "$asset_ready_stamp")"

legacy_entry="$(
  find "$asset_dir" -mindepth 1 -maxdepth 1 \
    ! -name '.keep' \
    ! -name "$(basename "$asset_archive")" \
    ! -name "$(basename "$asset_manifest")" \
    ! -name "$(basename "$client_support_registry")" \
    -print -quit
)"
if [[ -n "$legacy_entry" ]]; then
  printf 'embedded web asset directory contains legacy loose assets: %s\n' "$legacy_entry" >&2
  printf 'run make clean before rebuilding embedded web assets\n' >&2
  exit 2
fi

env GOCACHE="$go_cache_dir" GOMODCACHE="$go_mod_cache_dir" GOTMPDIR="$go_tmp_dir" \
  "$go_bin" run ./tools/embedwebassets \
  --source-dir "$(dirname "$source_index")" \
  --source-index "$source_index" \
  --output "$asset_archive" \
  --asset-manifest "$asset_manifest" \
  --client-support-source "$client_support_source" \
  --client-support-registry "$client_support_registry" \
  --temp-dir "$(dirname "$asset_stamp")"

write_stamp() {
  local stamp="$1"
  local tmp
  tmp="$(mktemp "$(dirname "$stamp")/.${stamp##*/}.tmp.XXXXXX")"
  printf 'source=%s\narchive=%s\nasset_manifest=%s\nclient_support_registry=%s\n' \
    "$source_index" "$asset_archive" "$asset_manifest" "$client_support_registry" >"$tmp"
  mv "$tmp" "$stamp"
}

write_stamp "$asset_ready_stamp"
write_stamp "$asset_stamp"
