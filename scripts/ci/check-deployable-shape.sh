#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)"
ALLOWED_RUNTIME_DIRS=("server" "migrate")

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

if [[ ! -f "apps/web/dist/index.html" ]]; then
  echo "deployable-shape check failed: frontend build artifact 'apps/web/dist/index.html' was not produced" >&2
  exit 1
fi

echo "deployable-shape verified: cmd/server remains the single runtime application unit and cmd/migrate remains operational tooling."
echo "hardening gap (explicit): frontend assets are still built separately and are not yet embedded into the server artifact."
