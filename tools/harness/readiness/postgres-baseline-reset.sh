#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
PROFILE="${POSTGRES_BASELINE_PROFILE:-dev}"
NODE_BIN="${NODE_BIN:-node}"

case "$PROFILE" in
  dev)
    COMPOSE_FILE="$ROOT_DIR/docker-compose.dev.yml"
    LEGACY_VOLUME_KEY="postgres-data"
    CURRENT_VOLUME_KEY="postgres-data-v18"
    ;;
  mvp)
    COMPOSE_FILE="$ROOT_DIR/deploy/mvp/docker-compose.yml"
    LEGACY_VOLUME_KEY="cartulary-postgres-data"
    CURRENT_VOLUME_KEY="cartulary-postgres-data-v18"
    ;;
  *)
    echo "postgres-baseline-reset: POSTGRES_BASELINE_PROFILE must be dev or mvp" >&2
    exit 2
    ;;
esac

compose() {
  docker compose -f "$COMPOSE_FILE" "$@"
}

compose_config="$(compose config --format json)"
project_name="$(printf '%s' "$compose_config" | "$NODE_BIN" -e '
let body="";
process.stdin.setEncoding("utf8");
process.stdin.on("data", chunk => body += chunk);
process.stdin.on("end", () => {
  const value = JSON.parse(body);
  if (typeof value.name !== "string" || !/^[a-z0-9][a-z0-9_-]*$/u.test(value.name)) process.exit(2);
  process.stdout.write(value.name);
});
')"
[[ -n "$project_name" ]] || { echo "postgres-baseline-reset: compose project identity is unavailable" >&2; exit 2; }

resolve_volume() {
  local logical_key="$1"
  local matches
  matches="$(docker volume ls --quiet \
    --filter "label=com.docker.compose.project=$project_name" \
    --filter "label=com.docker.compose.volume=$logical_key")"
  if [[ "$(printf '%s\n' "$matches" | sed '/^$/d' | wc -l)" -gt 1 ]]; then
    echo "postgres-baseline-reset: ambiguous volume for $project_name/$logical_key" >&2
    exit 1
  fi
  printf '%s' "$matches"
}

legacy_volume="$(resolve_volume "$LEGACY_VOLUME_KEY")"
current_volume="$(resolve_volume "$CURRENT_VOLUME_KEY")"

for tuple in "$LEGACY_VOLUME_KEY:$legacy_volume" "$CURRENT_VOLUME_KEY:$current_volume"; do
  logical_key="${tuple%%:*}"
  volume_name="${tuple#*:}"
  [[ -n "$volume_name" ]] || continue
  actual_project="$(docker volume inspect -f '{{index .Labels "com.docker.compose.project"}}' "$volume_name")"
  actual_key="$(docker volume inspect -f '{{index .Labels "com.docker.compose.volume"}}' "$volume_name")"
  if [[ "$actual_project" != "$project_name" || "$actual_key" != "$logical_key" ]]; then
    echo "postgres-baseline-reset: volume ownership proof failed for $volume_name" >&2
    exit 1
  fi
done

if [[ "${CARTULARY_CLEANUP_DRY_RUN:-}" == "1" ]]; then
  printf 'DRY-RUN stop-service compose:%s:postgres exact_compose_service\n' "$PROFILE"
  [[ -z "$legacy_volume" ]] || printf 'DRY-RUN remove-volume %s compose:%s/%s\n' "$legacy_volume" "$project_name" "$LEGACY_VOLUME_KEY"
  [[ -z "$current_volume" ]] || printf 'DRY-RUN remove-volume %s compose:%s/%s\n' "$current_volume" "$project_name" "$CURRENT_VOLUME_KEY"
  printf 'DRY-RUN provision-postgres compose:%s:postgres exact_postgresql_18_6\n' "$PROFILE"
  exit 0
fi

if [[ "${CARTULARY_DESTRUCTIVE_CONFIRM:-}" != "postgres-baseline-reset" ]]; then
  echo "refusing postgres-baseline-reset: set CARTULARY_DESTRUCTIVE_CONFIRM=postgres-baseline-reset or use CARTULARY_CLEANUP_DRY_RUN=1" >&2
  exit 2
fi

compose stop postgres >/dev/null 2>&1 || true
compose rm -f postgres >/dev/null 2>&1 || true
[[ -z "$legacy_volume" ]] || docker volume rm "$legacy_volume" >/dev/null
[[ -z "$current_volume" ]] || docker volume rm "$current_volume" >/dev/null
compose up -d postgres

deadline=$((SECONDS + 180))
until compose exec -T postgres pg_isready -U "${POSTGRES_USER:-cartulary}" -d "${POSTGRES_DB:-cartulary}" >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    echo "postgres-baseline-reset: PostgreSQL did not become ready" >&2
    compose logs --no-color --tail 120 postgres >&2 || true
    exit 1
  fi
  sleep 1
done

facts="$(compose exec -T postgres psql -U "${POSTGRES_USER:-cartulary}" -d "${POSTGRES_DB:-cartulary}" -Atc "SELECT current_setting('server_version_num'), current_setting('data_checksums');")"
[[ "$facts" == "180006|on" ]] || { echo "postgres-baseline-reset: PostgreSQL admission failed" >&2; exit 1; }

if [[ "$PROFILE" == "dev" ]]; then
  env CARTULARY_DESTRUCTIVE_CONFIRM=db-reset \
    bash "$ROOT_DIR/tools/harness/readiness/dev-services.sh" db-reset
else
  compose run --rm migrate
fi

echo "postgres-baseline-reset: $PROFILE now uses exact PostgreSQL 18.6 with checksums on"
