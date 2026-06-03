#!/usr/bin/env bash
set -euo pipefail

source_index="${WEB_DIST_INDEX:?WEB_DIST_INDEX is required}"
asset_dir="${EMBEDDED_WEB_ASSET_DIR:?EMBEDDED_WEB_ASSET_DIR is required}"
asset_stamp="${EMBEDDED_WEB_ASSET_STAMP:?EMBEDDED_WEB_ASSET_STAMP is required}"
asset_ready_stamp="${EMBEDDED_WEB_ASSET_READY_STAMP:?EMBEDDED_WEB_ASSET_READY_STAMP is required}"

mkdir -p "$asset_dir" "$(dirname "$asset_stamp")"
find "$asset_dir" -mindepth 1 ! -name '.keep' -exec rm -rf {} +
cp -R "$(dirname "$source_index")/." "$asset_dir/"
printf 'source=%s\n' "$source_index" >"$asset_ready_stamp"
printf 'source=%s\n' "$source_index" >"$asset_stamp"
