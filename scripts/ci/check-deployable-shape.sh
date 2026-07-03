#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../.." && pwd)"
ALLOWED_RUNTIME_DIRS=("server" "migrate" "operator")
EMBEDDED_WEB_DIR="internal/platform/httpapi/webassets/dist"
EMBEDDED_WEB_ARCHIVE="${EMBEDDED_WEB_DIR}/web-assets.zip"

cd "$ROOT_DIR"

fail() {
  echo "deployable-shape check failed: $*" >&2
  exit 1
}

if [[ ! -f "cmd/server/main.go" ]]; then
  fail "missing required runtime entrypoint cmd/server/main.go"
fi

if [[ ! -f "cmd/migrate/main.go" ]]; then
  fail "missing required operational entrypoint cmd/migrate/main.go"
fi

if [[ ! -f "cmd/operator/main.go" ]]; then
  fail "missing required operational entrypoint cmd/operator/main.go"
fi

mapfile -t MAIN_ENTRYPOINTS < <(find cmd -mindepth 2 -maxdepth 2 -type f -name 'main.go' | LC_ALL=C sort)
if [[ "${#MAIN_ENTRYPOINTS[@]}" -ne 3 ]]; then
  echo "deployable-shape check failed: expected exactly three cmd/*/main.go entrypoints (server + migrate + operator)" >&2
  printf 'found entrypoints:\n' >&2
  printf '  %s\n' "${MAIN_ENTRYPOINTS[@]}" >&2
  exit 1
fi

for main_file in "${MAIN_ENTRYPOINTS[@]}"; do
  entry_dir="$(basename "$(dirname "$main_file")")"
  case "$entry_dir" in
    "${ALLOWED_RUNTIME_DIRS[0]}"|"${ALLOWED_RUNTIME_DIRS[1]}"|"${ALLOWED_RUNTIME_DIRS[2]}")
      ;;
    *)
      fail "unexpected deployable entrypoint $main_file"
      ;;
  esac
done

if [[ ! -f "server" ]]; then
  fail "backend build artifact './server' was not produced"
fi

if [[ ! -f "migrate" ]]; then
  fail "migration build artifact './migrate' was not produced"
fi

if [[ ! -f "operator" ]]; then
  fail "operator build artifact './operator' was not produced"
fi

if [[ ! -f "${EMBEDDED_WEB_ARCHIVE}" ]]; then
  fail "embedded frontend archive '${EMBEDDED_WEB_ARCHIVE}' was not produced"
fi

if ! command -v zipinfo >/dev/null 2>&1; then
  fail "zipinfo is required to inspect embedded frontend archive '${EMBEDDED_WEB_ARCHIVE}'"
fi

if ! command -v unzip >/dev/null 2>&1; then
  fail "unzip is required to inspect embedded frontend archive '${EMBEDDED_WEB_ARCHIVE}'"
fi

embedded_index="$(mktemp)"
cleanup() {
  rm -f "$embedded_index"
}
trap cleanup EXIT

zip_listing="$(zipinfo -1 "${EMBEDDED_WEB_ARCHIVE}")"
if ! printf '%s\n' "$zip_listing" | grep -Fxq 'index.html'; then
  fail "embedded frontend archive '${EMBEDDED_WEB_ARCHIVE}' is missing index.html"
fi

if ! unzip -p "${EMBEDDED_WEB_ARCHIVE}" index.html >"$embedded_index"; then
  fail "embedded frontend archive '${EMBEDDED_WEB_ARCHIVE}' contains an unreadable index.html"
fi

if ! grep -Fq '<div id="root"></div>' "$embedded_index"; then
  fail "embedded frontend index.html is missing the application root shell"
fi

first_embedded_asset="$(
  grep -Eo '(src|href)="[^"]+"' "$embedded_index" \
    | sed -n 's#.*="/\{0,1\}\([^"?]*assets/[^"?]*\)".*#\1#p' \
    | LC_ALL=C sort \
    | head -n 1
)"
if [[ -z "${first_embedded_asset}" ]]; then
  fail "embedded frontend index.html does not reference a built asset"
fi
if ! printf '%s\n' "$zip_listing" | grep -Fxq "$first_embedded_asset"; then
  fail "embedded frontend archive '${EMBEDDED_WEB_ARCHIVE}' is missing referenced asset '${first_embedded_asset}'"
fi

embedded_asset_name="$(basename "${first_embedded_asset}")"
if ! grep -aFq '<div id="root"></div>' "server"; then
  fail "backend build artifact './server' does not appear to embed the frontend root shell"
fi
if ! grep -aFq "${embedded_asset_name}" "server"; then
  fail "backend build artifact './server' does not appear to embed frontend asset '${embedded_asset_name}'"
fi

echo "deployable-shape verified: cmd/server remains the single runtime application unit, cmd/migrate and cmd/operator remain operational tooling, and the built server binary embeds the frontend archive."
