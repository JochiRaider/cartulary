#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)"
ALLOWED_RUNTIME_DIRS=("server" "migrate")
EMBEDDED_WEB_DIR="internal/platform/httpapi/webassets/dist"

cd "$ROOT_DIR"

if [[ ! -f "cmd/server/main.go" ]]; then
  echo "deployable-shape check failed: missing required runtime entrypoint cmd/server/main.go" >&2
  exit 1
fi

if [[ ! -f "cmd/migrate/main.go" ]]; then
  echo "deployable-shape check failed: missing required operational entrypoint cmd/migrate/main.go" >&2
  exit 1
fi

mapfile -t MAIN_ENTRYPOINTS < <(find cmd -mindepth 2 -maxdepth 2 -type f -name 'main.go' | LC_ALL=C sort)
if [[ "${#MAIN_ENTRYPOINTS[@]}" -ne 2 ]]; then
  echo "deployable-shape check failed: expected exactly two cmd/*/main.go entrypoints (server + migrate)" >&2
  printf 'found entrypoints:\n' >&2
  printf '  %s\n' "${MAIN_ENTRYPOINTS[@]}" >&2
  exit 1
fi

for main_file in "${MAIN_ENTRYPOINTS[@]}"; do
  entry_dir="$(basename "$(dirname "$main_file")")"
  case "$entry_dir" in
    "${ALLOWED_RUNTIME_DIRS[0]}"|"${ALLOWED_RUNTIME_DIRS[1]}")
      ;;
    *)
      echo "deployable-shape check failed: unexpected deployable entrypoint $main_file" >&2
      exit 1
      ;;
  esac
done

if [[ ! -f "server" ]]; then
  echo "deployable-shape check failed: backend build artifact './server' was not produced" >&2
  exit 1
fi

if [[ ! -f "migrate" ]]; then
  echo "deployable-shape check failed: migration build artifact './migrate' was not produced" >&2
  exit 1
fi

if [[ ! -f "${EMBEDDED_WEB_DIR}/index.html" ]]; then
  echo "deployable-shape check failed: embedded frontend asset '${EMBEDDED_WEB_DIR}/index.html' was not produced" >&2
  exit 1
fi

if ! grep -Fq '<div id="root"></div>' "${EMBEDDED_WEB_DIR}/index.html"; then
  echo "deployable-shape check failed: embedded frontend index.html is missing the application root shell" >&2
  exit 1
fi

first_embedded_asset="$(find "${EMBEDDED_WEB_DIR}/assets" -type f | LC_ALL=C sort | head -n 1)"
if [[ -z "${first_embedded_asset}" ]]; then
  echo "deployable-shape check failed: embedded frontend assets directory '${EMBEDDED_WEB_DIR}/assets' is empty" >&2
  exit 1
fi

embedded_asset_name="$(basename "${first_embedded_asset}")"
if ! grep -aFq '<div id="root"></div>' "server"; then
  echo "deployable-shape check failed: backend build artifact './server' does not appear to embed the frontend root shell" >&2
  exit 1
fi
if ! grep -aFq "${embedded_asset_name}" "server"; then
  echo "deployable-shape check failed: backend build artifact './server' does not appear to embed frontend asset '${embedded_asset_name}'" >&2
  exit 1
fi

echo "deployable-shape verified: cmd/server remains the single runtime application unit, cmd/migrate remains operational tooling, and the built server binary embeds the frontend app."
