#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../.." && pwd)"
PACKAGE_DIR="$ROOT_DIR/deploy/mvp"
PUBLIC_ORIGIN_PATH="/ws/v1/incidents/00000000-0000-0000-0000-000000000000"

require_command() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "standup-package-smoke failed: missing required command $name" >&2
    exit 2
  fi
}

fail() {
  echo "standup-package-smoke failed: $*" >&2
  exit 1
}

pick_port() {
  local port
  for port in $(seq 39000 39999 | sort -R | head -n 200); do
    if ! (echo >"/dev/tcp/127.0.0.1/${port}") >/dev/null 2>&1; then
      printf '%s\n' "$port"
      return 0
    fi
  done
  return 1
}

require_command docker
require_command curl
require_command grep
require_command sed
require_command sort
require_command tar

docker info >/dev/null 2>&1 || fail "docker daemon is not available"
docker compose version >/dev/null 2>&1 || fail "docker compose plugin is not available"

work_dir="$(mktemp -d "${TMPDIR:-/tmp}/cartulary-standup-package-smoke.XXXXXX")"
project="cartularymvpsmoke$(date +%s)$$"
image="cartulary/mvp-smoke:${project}"
port="$(pick_port)" || fail "no free loopback port found for package smoke"
public_origin="http://127.0.0.1:${port}"
compose_file="$work_dir/docker-compose.yml"
image_listing="$work_dir/image-files.txt"
root_body="$work_dir/root.html"
asset_body="$work_dir/asset.bin"
ready_body="$work_dir/readyz.json"
container_id=""
legacy_postgres_volume="${project}_legacy-postgres-v16"
postgres_image="docker.io/library/postgres:18.6-alpine@sha256:d3e1620b530c944afa6e887d22eb899824da68e19c52024bf98f5220c88a65b2"

compose_overlay=()
compose() {
  docker compose --project-name "$project" --env-file "$work_dir/.env" -f "$compose_file" "${compose_overlay[@]}" "$@"
}

# shellcheck source=tools/release-evidence/package-smoke-cleanup.sh
source "$ROOT_DIR/tools/release-evidence/package-smoke-cleanup.sh"
qualification="${CARTULARY_PACKAGE_QUALIFICATION:-package}"
case "$qualification" in package|reference-pack) ;; *) fail "unsupported package qualification" ;; esac
artifact_dir="${CARTULARY_TEST_RESULTS_DIR:-${ROOT_DIR}/.cartulary/test-results}/${CARTULARY_TEST_RUN_ID:-standup-package-smoke-manual}/standup-${qualification}-smoke/artifacts"
mkdir -p "$artifact_dir"
cleanup() {
  local status=$?
  trap - EXIT INT TERM
  set +e
  if ! cleanup_package_workspace "$project" "$work_dir" "$postgres_image" "$artifact_dir/workspace-cleanup.json"; then status=1; fi
  if ! cleanup_package_resources "${project}destination" "" "$artifact_dir/destination-cleanup.json"; then status=1; fi
  if ! cleanup_package_resources "$project" "$image" "$artifact_dir/cleanup.json"; then status=1; fi
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

sed "s#context: ../..#context: ${ROOT_DIR}#g" "$PACKAGE_DIR/docker-compose.yml" >"$compose_file"
cp "$PACKAGE_DIR/config.toml.example" "$work_dir/config.toml"
cp "$PACKAGE_DIR/bootstrap-admin.json.example" "$work_dir/bootstrap-admin.json"
cp "$PACKAGE_DIR/revisions-conflict-token-key-ring.json.example" "$work_dir/revisions-conflict-token-key-ring.json"
cp "$PACKAGE_DIR/postgres-provision.sh" "$work_dir/postgres-provision.sh"
chmod 0755 "$work_dir/postgres-provision.sh"
chmod 0644 "$work_dir/config.toml" "$work_dir/bootstrap-admin.json" "$work_dir/revisions-conflict-token-key-ring.json"
cat >"$work_dir/.env" <<EOF
CARTULARY_IMAGE=${image}
CARTULARY_HTTP_PORT=${port}
CARTULARY_PUBLIC_ORIGIN=${public_origin}

POSTGRES_DB=cartulary
POSTGRES_USER=cartulary
POSTGRES_PASSWORD=cartulary-postgres-smoke-password
CARTULARY_POSTGRES_MIGRATION_PASSWORD=cartulary-migration-smoke-password
CARTULARY_POSTGRES_RUNTIME_PASSWORD=cartulary-runtime-smoke-password
CARTULARY_POSTGRES_RECOVERY_PASSWORD=cartulary-recovery-smoke-password

CARTULARY_AUTH_MASTER_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=
CARTULARY_RECOVERY_MASTER_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=
CARTULARY_SECRET_REVISIONS_CONFLICT_TOKEN_ACTIVE=cmV2aXNpb25zLXRva2VuLWtleS1tYXRlcmlhbC0wMDE

CARTULARY_S3_PRIMARY_ACCESS_KEY_ID=cartulary-local
CARTULARY_S3_PRIMARY_SECRET_ACCESS_KEY=cartulary-local-secret
CARTULARY_S3_PRIMARY_BUCKET=cartulary-mvp-smoke
CARTULARY_S3_RESTORE_VERIFY_ENDPOINT=seaweedfs-s3:8333
CARTULARY_S3_RESTORE_VERIFY_ACCESS_KEY_ID=cartulary-local
CARTULARY_S3_RESTORE_VERIFY_SECRET_ACCESS_KEY=cartulary-local-secret
CARTULARY_S3_RESTORE_VERIFY_BUCKET=cartulary-mvp-smoke-restore
CARTULARY_S3_RESTORE_VERIFY_SECURE=false
EOF

compose build app >/dev/null

container_id="$(docker create --label "com.docker.compose.project=${project}" "$image")"
docker export "$container_id" | tar -tf - >"$image_listing"
docker rm "$container_id" >/dev/null
container_id=""

for required in \
  "usr/local/bin/cartulary-server" \
  "usr/local/bin/cartulary-migrate" \
  "usr/local/bin/cartulary-operator"; do
  grep -Fxq "$required" "$image_listing" || fail "image is missing $required"
done

if grep -Eq '(^|/)(node|npm|pnpm|vite)(/|$)|(^|/)node_modules(/|$)|(^|/)apps/web(/|$)|(^|/)db/migrations(/|$)' "$image_listing"; then
  fail "runtime image contains forbidden Node/Vite/source-tree paths"
fi

docker volume create --label "com.docker.compose.project=${project}" --label "com.docker.compose.volume=legacy-postgres-v16" "$legacy_postgres_volume" >/dev/null
docker run --rm --label "com.docker.compose.project=${project}" --entrypoint sh -v "${legacy_postgres_volume}:/var/lib/postgresql" "$postgres_image" -c \
  'mkdir -p /var/lib/postgresql/18/docker && printf "16\n" >/var/lib/postgresql/18/docker/PG_VERSION'
if docker run --rm --label "com.docker.compose.project=${project}" \
  -e PGDATA=/var/lib/postgresql/18/docker \
  -e POSTGRES_PASSWORD=unused-legacy-fixture \
  -v "${legacy_postgres_volume}:/var/lib/postgresql" \
  "$postgres_image" >"$work_dir/legacy-postgres.stdout" 2>"$work_dir/legacy-postgres.stderr"; then
  fail "PostgreSQL 18 accepted a synthetic PostgreSQL 16 data directory"
fi
docker volume rm "$legacy_postgres_volume" >/dev/null

compose up -d --build app >/dev/null

wait_for_http_status() {
  local path="$1"
  local want="$2"
  local output="$3"
  local start_time="$SECONDS"
  local status="000"
  while ((SECONDS - start_time < 180)); do
    status="$(curl -sS -o "$output" -w '%{http_code}' "${public_origin}${path}" || true)"
    if [[ "$status" == "$want" ]]; then
      return 0
    fi
    sleep 1
  done
  echo "last status for ${path}: ${status}" >&2
  return 1
}

wait_for_http_status "/healthz" "200" "$work_dir/healthz.txt" || fail "/healthz did not become live"
wait_for_http_status "/readyz" "200" "$ready_body" || fail "/readyz did not become ready"
grep -Fq '"status":"ready"' "$ready_body" || fail "/readyz did not report structured ready status"

wait_for_http_status "/" "200" "$root_body" || fail "embedded frontend root did not load"
grep -Fq '<div id="root"></div>' "$root_body" || fail "embedded root shell is missing"
if grep -Fq '@vite/client' "$root_body"; then
  fail "embedded root references Vite dev client"
fi
asset_path="$(grep -Eo '"/assets/[^"]+"' "$root_body" | head -n 1 | tr -d '"')"
if [[ -z "$asset_path" ]]; then
  fail "embedded root did not reference a packaged asset"
fi
curl -fsS "${public_origin}${asset_path}" -o "$asset_body" || fail "embedded asset ${asset_path} did not load"
[[ -s "$asset_body" ]] || fail "embedded asset ${asset_path} was empty"

assert_completed_zero() {
  local service="$1"
  local id
  id="$(compose ps -a -q "$service")"
  [[ -n "$id" ]] || fail "missing compose container for $service"
  local exit_code
  exit_code="$(docker inspect -f '{{.State.ExitCode}}' "$id")"
  [[ "$exit_code" == "0" ]] || fail "$service exited with $exit_code"
}

assert_completed_zero migrate
assert_completed_zero object-store-init

migration_count="$(compose exec -T postgres psql -U cartulary -d cartulary -Atc "SELECT COUNT(*) FROM goose_db_version WHERE is_applied = true;" | tr -d '[:space:]')"
[[ "$migration_count" =~ ^[0-9]+$ ]] || fail "migration count was not numeric"
((migration_count > 0)) || fail "migrate did not apply migrations"

assert_volume_mount() {
  local service="$1"
  local destination="$2"
  local id
  id="$(compose ps -q "$service")"
  [[ -n "$id" ]] || fail "missing running compose container for $service"
  if ! docker inspect -f '{{range .Mounts}}{{println .Type "|" .Source "|" .Destination}}{{end}}' "$id" | awk -F '|' -v dest="$destination" '$1 == "volume " && $3 == " " dest { found=1 } END { exit found ? 0 : 1 }'; then
    fail "$service does not mount $destination as a Docker volume"
  fi
}

assert_bind_mount_not_from_source_tree() {
  local service="$1"
  local id
  id="$(compose ps -q "$service")"
  [[ -n "$id" ]] || fail "missing running compose container for $service"
  while IFS='|' read -r mount_type source destination; do
    mount_type="${mount_type%% }"
    source="${source# }"
    source="${source% }"
    destination="${destination# }"
    if [[ "$mount_type" == "bind" && "$source" == "$ROOT_DIR"* ]]; then
      fail "$service bind-mounts source-tree path $source to $destination"
    fi
  done < <(docker inspect -f '{{range .Mounts}}{{println .Type "|" .Source "|" .Destination}}{{end}}' "$id")
}

assert_volume_mount postgres /var/lib/postgresql
# shellcheck disable=SC2016 # Expand PGDATA inside the Compose container, not this shell.
effective_pgdata="$(compose exec -T postgres sh -c 'printf %s "$PGDATA"')"
[[ "$effective_pgdata" == "/var/lib/postgresql/18/docker" ]] || fail "postgres PGDATA was $effective_pgdata"
server_facts="$(compose exec -T postgres psql -U cartulary -d cartulary -Atc "SELECT current_setting('server_version_num'), current_setting('data_checksums');")"
[[ "$server_facts" == "180006|on" ]] || fail "postgres engine admission facts were not exact 18.6/checksums-on"
assert_volume_mount seaweedfs-s3 /data
assert_volume_mount app /var/lib/cartulary/backups
assert_volume_mount app /var/lib/cartulary/reference-packs
assert_volume_mount app /var/lib/cartulary/tmp
assert_volume_mount app /var/lib/cartulary/exports
assert_bind_mount_not_from_source_tree app

ws_common=(
  -H "Connection: Upgrade"
  -H "Upgrade: websocket"
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ=="
  -H "Sec-WebSocket-Version: 13"
  -H "Cookie: cartulary_session=dummy-session"
)
untrusted_status="$(curl -sS -o "$work_dir/ws-untrusted.txt" -w '%{http_code}' "${ws_common[@]}" -H "Origin: https://untrusted.example.test" "${public_origin}${PUBLIC_ORIGIN_PATH}" || true)"
[[ "$untrusted_status" == "403" ]] || fail "untrusted cookie-authenticated WebSocket Origin returned $untrusted_status, want 403"
trusted_status="$(curl -sS -o "$work_dir/ws-trusted.txt" -w '%{http_code}' "${ws_common[@]}" -H "Origin: ${public_origin}" "${public_origin}${PUBLIC_ORIGIN_PATH}" || true)"
[[ "$trusted_status" != "403" ]] || fail "configured WebSocket Origin was rejected with 403"

# Qualify the optional shipped administration setup against the same binaries.
cp "$PACKAGE_DIR/docker-compose.reference-packs.yml" "$work_dir/reference-packs.yml"
"${NODE_BIN:-node}" --input-type=module - "$ROOT_DIR" "$work_dir/reference-pack-incoming" <<'EOF'
import { pathToFileURL } from "node:url";
const [root, destination] = process.argv.slice(2);
const { packageFixtures } = await import(pathToFileURL(`${root}/tools/release-evidence/reference-pack-fixtures.mjs`));
packageFixtures(root, destination);
EOF
chmod 0755 "$work_dir/reference-pack-incoming"
chmod 0644 "$work_dir/reference-pack-incoming/"*.tar
compose stop app >/dev/null
cat "$PACKAGE_DIR/reference-pack-administration.toml.example" >>"$work_dir/config.toml"
compose_overlay=(-f "$work_dir/reference-packs.yml")
if compose run --rm --no-deps app >"$work_dir/missing-trust.log" 2>&1; then
  fail "claimed package accepted a missing trust mount"
fi
# Some Compose run versions materialize a missing bind as an empty directory
# despite create_host_path=false. Admission must still fail; remove only that
# empty test-owned path before the malformed-file case.
if [[ -d "$work_dir/reference-pack-trust.json" ]]; then
  rmdir "$work_dir/reference-pack-trust.json"
fi
printf '{}\n' >"$work_dir/reference-pack-trust.json"
chmod 0644 "$work_dir/reference-pack-trust.json"
if compose run --rm --no-deps app >"$work_dir/invalid-trust.log" 2>&1; then
  fail "claimed package accepted invalid trust"
fi
cp "$work_dir/reference-pack-incoming/trust.json" "$work_dir/reference-pack-trust.json"
compose up -d --no-deps --force-recreate app >/dev/null
wait_for_http_status "/readyz" "200" "$ready_body" || fail "claimed package did not become ready"
if compose run --rm --no-deps reference-pack-operator reference-pack import valid.tar >"$work_dir/untrusted-clock.json"; then
  fail "operator accepted an untrusted clock"
fi
grep -Fq 'clock_untrusted' "$work_dir/untrusted-clock.json" || fail "operator did not report clock rejection"
compose stop app >/dev/null
sed 's/clock_trusted = false/clock_trusted = true/' "$work_dir/config.toml" >"$work_dir/clock.toml"
cat "$work_dir/clock.toml" >"$work_dir/config.toml"
compose up -d --no-deps --force-recreate app >/dev/null
wait_for_http_status "/readyz" "200" "$ready_body" || fail "trusted claimed package did not become ready"
compose run --rm --no-deps reference-pack-operator reference-pack import valid.tar >"$work_dir/import.json"
compose run --rm --no-deps reference-pack-operator reference-pack import valid.tar >"$work_dir/replay.json"
"${NODE_BIN:-node}" - "$work_dir/import.json" "$work_dir/replay.json" <<'EOF'
const fs = require("node:fs");
const [first, replay] = process.argv.slice(2).map((file) => JSON.parse(fs.readFileSync(file)));
if (first.schema_id !== "cartulary.reference_pack_operator_result.v1" || first.result !== "succeeded" || replay.result !== "succeeded" || first.container_sha256 !== replay.container_sha256 || first.pack_key !== "enrichment.tor") throw new Error("packaged operator import/replay identity mismatch");
EOF
docker image inspect --format '{{.Id}}' "$image" >"$artifact_dir/image-id.txt"
sha256sum "$work_dir/config.toml" "$PACKAGE_DIR/docker-compose.reference-packs.yml" | sed "s#$work_dir/##;s#$ROOT_DIR/##" >"$artifact_dir/configuration-digests.txt"
cp "$work_dir/import.json" "$artifact_dir/reference-pack-import.json"

if [[ "$qualification" == reference-pack ]]; then
  "${NODE_BIN:-node}" "$ROOT_DIR/tools/release-evidence/standup-reference-pack-scenarios.mjs" "$ROOT_DIR" "$work_dir" "$project" "$public_origin" "$artifact_dir"
fi

echo "standup-package-smoke verified: image shape, compose runtime, migrations, object-store init, embedded assets, readiness, persistent roots, and WebSocket Origin behavior."
