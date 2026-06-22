#!/usr/bin/env bash
set -euo pipefail

source_index="${WEB_DIST_INDEX:?WEB_DIST_INDEX is required}"
asset_dir="${EMBEDDED_WEB_ASSET_DIR:?EMBEDDED_WEB_ASSET_DIR is required}"
asset_archive="${EMBEDDED_WEB_ASSET_ARCHIVE:?EMBEDDED_WEB_ASSET_ARCHIVE is required}"
asset_stamp="${EMBEDDED_WEB_ASSET_STAMP:?EMBEDDED_WEB_ASSET_STAMP is required}"
asset_ready_stamp="${EMBEDDED_WEB_ASSET_READY_STAMP:?EMBEDDED_WEB_ASSET_READY_STAMP is required}"
go_bin="${GO:?GO is required}"

mkdir -p "$asset_dir" "$(dirname "$asset_stamp")" "$(dirname "$asset_ready_stamp")"

legacy_entry="$(
  find "$asset_dir" -mindepth 1 -maxdepth 1 \
    ! -name '.keep' \
    ! -name "$(basename "$asset_archive")" \
    -print -quit
)"
if [[ -n "$legacy_entry" ]]; then
  printf 'embedded web asset directory contains legacy loose assets: %s\n' "$legacy_entry" >&2
  printf 'run make clean before rebuilding embedded web assets\n' >&2
  exit 2
fi

"$go_bin" run ./tools/embedwebassets \
  --source-dir "$(dirname "$source_index")" \
  --source-index "$source_index" \
  --output "$asset_archive" \
  --temp-dir "$(dirname "$asset_stamp")"

write_stamp() {
  local stamp="$1"
  local tmp
  tmp="$(mktemp "$(dirname "$stamp")/.${stamp##*/}.tmp.XXXXXX")"
  printf 'source=%s\narchive=%s\n' "$source_index" "$asset_archive" >"$tmp"
  mv "$tmp" "$stamp"
}

write_stamp "$asset_ready_stamp"
write_stamp "$asset_stamp"
