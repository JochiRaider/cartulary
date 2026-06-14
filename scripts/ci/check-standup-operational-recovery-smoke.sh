#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../.." && pwd)"
PACKAGE_DIR="$ROOT_DIR/deploy/mvp"
NODE="${NODE_BIN:-node}"
PUBLIC_ORIGIN_PATH="/ws/v1/incidents/00000000-0000-0000-0000-000000000000"

require_command() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "standup-operational-recovery-smoke failed: missing required command $name" >&2
    exit 2
  fi
}

fail() {
  echo "standup-operational-recovery-smoke failed: $*" >&2
  exit 1
}

pick_port() {
  local port
  for port in $(seq 40000 40999 | sort -R | head -n 200); do
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
require_command "$NODE"

docker info >/dev/null 2>&1 || fail "docker daemon is not available"
docker compose version >/dev/null 2>&1 || fail "docker compose plugin is not available"

work_dir="$(mktemp -d "${TMPDIR:-/tmp}/cartulary-standup-recovery-smoke.XXXXXX")"
project="cartularymvprecoverysmk$(date +%s)$$"
image="cartulary/mvp-recovery-smoke:${project}"
port="$(pick_port)" || fail "no free loopback port found for operational recovery smoke"
public_origin="http://127.0.0.1:${port}"
compose_file="$work_dir/docker-compose.yml"
capture_json="$work_dir/backup-capture.json"
latest_json="$work_dir/latest-backup.json"
restore_verify_json="$work_dir/restore-verify-due.json"
route_json="$work_dir/public-route-absence.json"
summary_json="$work_dir/standup-operational-recovery-summary.json"
ready_body="$work_dir/readyz.json"

results_root="${CARTULARY_TEST_RESULTS_DIR:-${ROOT_DIR}/.cartulary/test-results}"
run_id="${CARTULARY_TEST_RUN_ID:-standup-operational-recovery-smoke-manual}"
artifact_dir="${results_root}/${run_id}/standup-operational-recovery-smoke/artifacts"
mkdir -p "$artifact_dir"

compose() {
  docker compose --project-name "$project" --env-file "$work_dir/.env" -f "$compose_file" "$@"
}

dump_compose_diagnostics() {
  compose ps -a >&2 || true
  compose logs --no-color --tail 200 >&2 || true
}

cleanup() {
  local status=$?
  set +e
  if [[ "$status" -ne 0 ]]; then
    dump_compose_diagnostics
  fi
  compose down -v --remove-orphans >/dev/null 2>&1 || true
  docker rmi "$image" >/dev/null 2>&1 || true
  cp -f "$capture_json" "$artifact_dir/backup-capture.json" 2>/dev/null || true
  cp -f "$latest_json" "$artifact_dir/latest-backup.json" 2>/dev/null || true
  cp -f "$restore_verify_json" "$artifact_dir/restore-verify-due.json" 2>/dev/null || true
  cp -f "$route_json" "$artifact_dir/public-route-absence.json" 2>/dev/null || true
  cp -f "$summary_json" "$artifact_dir/standup-operational-recovery-summary.json" 2>/dev/null || true
  rm -rf "$work_dir"
  exit "$status"
}
trap cleanup EXIT

sed "s#context: ../..#context: ${ROOT_DIR}#g" "$PACKAGE_DIR/docker-compose.yml" >"$compose_file"
cp "$PACKAGE_DIR/config.toml.example" "$work_dir/config.toml"
cp "$PACKAGE_DIR/bootstrap-admin.json.example" "$work_dir/bootstrap-admin.json"
cp "$PACKAGE_DIR/restore-verification-target.toml.example" "$work_dir/restore-verification-target.toml"
cp "$PACKAGE_DIR/restore-verification-target.marker.json.example" "$work_dir/restore-verification-target.marker.json.example"
mkdir -p "$work_dir/runtime/restore-verification-target"

cat >"$work_dir/.env" <<EOF
CARTULARY_IMAGE=${image}
CARTULARY_HTTP_PORT=${port}
CARTULARY_PUBLIC_ORIGIN=${public_origin}

POSTGRES_DB=cartulary
POSTGRES_USER=cartulary
POSTGRES_PASSWORD=cartulary-postgres-smoke-password

CARTULARY_AUTH_MASTER_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=
CARTULARY_RECOVERY_MASTER_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=
CARTULARY_RECOVERY_DEPLOYMENT_ADMIN_EMAIL=admin@example.test

CARTULARY_S3_PRIMARY_ACCESS_KEY_ID=cartulary-local
CARTULARY_S3_PRIMARY_SECRET_ACCESS_KEY=cartulary-local-secret
CARTULARY_S3_PRIMARY_BUCKET=cartulary-mvp-smoke

RESTORE_VERIFY_POSTGRES_DB=cartulary_restore_verify
CARTULARY_POSTGRES_RESTORE_VERIFY_DSN=postgresql://cartulary:cartulary-postgres-smoke-password@postgres:5432/cartulary_restore_verify?sslmode=disable
CARTULARY_S3_RESTORE_VERIFY_ENDPOINT=seaweedfs-s3:8333
CARTULARY_S3_RESTORE_VERIFY_ACCESS_KEY_ID=cartulary-local
CARTULARY_S3_RESTORE_VERIFY_SECRET_ACCESS_KEY=cartulary-local-secret
CARTULARY_S3_RESTORE_VERIFY_BUCKET=cartulary-mvp-restore-verify-smoke
CARTULARY_S3_RESTORE_VERIFY_SECURE=false
EOF

compose build app >/dev/null
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

wait_for_http_status "/readyz" "200" "$ready_body" || fail "/readyz did not become ready"
grep -Fq '"status":"ready"' "$ready_body" || fail "/readyz did not report structured ready status"

CARTULARY_MVP_DIR="$work_dir" \
  CARTULARY_MVP_ENV_FILE="$work_dir/.env" \
  CARTULARY_MVP_COMPOSE_FILE="$compose_file" \
  CARTULARY_MVP_COMPOSE_PROJECT_NAME="$project" \
  CARTULARY_SOURCE_CONFIG="$work_dir/config.toml" \
  "$PACKAGE_DIR/scripts/backup-capture.sh" >"$capture_json"

compose run --rm --no-deps \
  --entrypoint /usr/local/bin/cartulary-operator \
  app backup-metadata latest \
  -source-config /etc/cartulary/config.toml \
  -deployment-admin-email admin@example.test >"$latest_json"

"$NODE" - "$capture_json" "$latest_json" <<'EOF'
const fs = require("node:fs");
const capture = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const latest = JSON.parse(fs.readFileSync(process.argv[3], "utf8"));
const dayMs = 24 * 60 * 60 * 1000;
function fail(message) {
  console.error(message);
  process.exit(1);
}
if (capture.schema_id !== "cartulary.operator.backup_capture_result.v1") {
  fail(`unexpected capture schema_id ${capture.schema_id}`);
}
if (latest.schema_id !== "cartulary.operator.backup_metadata.v1") {
  fail(`unexpected latest schema_id ${latest.schema_id}`);
}
if (capture.backup_set_id !== latest.backup_set_id) {
  fail("latest metadata did not select captured backup_set");
}
const consistency = new Date(latest.consistency_point_at).getTime();
if (!Number.isFinite(consistency) || Date.now() - consistency > dayMs) {
  fail("latest backup consistency point is older than 24 hours");
}
const created = new Date(latest.created_at).getTime();
const retained = new Date(latest.retained_until).getTime();
if (!Number.isFinite(created) || !Number.isFinite(retained) || retained - created < 30 * dayMs) {
  fail("latest backup retention is below 30 days");
}
for (const field of [
  "postgres_artifact_key",
  "postgres_artifact_sha256",
  "object_store_artifact_key",
  "object_store_artifact_sha256",
  "integrity_manifest_key",
  "integrity_manifest_sha256",
]) {
  if (typeof latest[field] !== "string" || latest[field].length === 0) {
    fail(`latest backup missing ${field}`);
  }
}
EOF

CARTULARY_MVP_DIR="$work_dir" \
  CARTULARY_MVP_ENV_FILE="$work_dir/.env" \
  CARTULARY_MVP_COMPOSE_FILE="$compose_file" \
  CARTULARY_MVP_COMPOSE_PROJECT_NAME="$project" \
  CARTULARY_SOURCE_CONFIG="$work_dir/config.toml" \
  CARTULARY_RESTORE_VERIFY_TARGET_CONFIG="$work_dir/restore-verification-target.toml" \
  CARTULARY_RESTORE_VERIFY_TARGET_ROOT="$work_dir/runtime/restore-verification-target" \
  "$PACKAGE_DIR/scripts/restore-verify-due.sh" >"$restore_verify_json"

"$NODE" - "$restore_verify_json" <<'EOF'
const fs = require("node:fs");
const due = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
function fail(message) {
  console.error(message);
  process.exit(1);
}
if (due.schema_id !== "cartulary.operator.restore_verification_due_result.v1") {
  fail(`unexpected due schema_id ${due.schema_id}`);
}
if (due.due_count < 1 || due.verified_count < 1 || due.failed_count !== 0) {
  fail(`unexpected due verification counts ${JSON.stringify(due)}`);
}
for (const item of due.results ?? []) {
  if (item.verification_state !== "verified") {
    fail(`restore verification item was not verified: ${JSON.stringify(item)}`);
  }
  for (const field of [
    "restore_verification_artifact_key",
    "restore_verification_artifact_sha256",
  ]) {
    if (typeof item[field] !== "string" || item[field].length === 0) {
      fail(`restore verification item missing ${field}`);
    }
  }
  if (!(item.restore_verification_artifact_size_bytes > 0)) {
    fail("restore verification item missing positive artifact size");
  }
}
EOF

"$NODE" - "$public_origin" "$route_json" <<'EOF'
const fs = require("node:fs");
const http = require("node:http");
const base = new URL(process.argv[2]);
const output = process.argv[3];
const httpPaths = ["/api/v1/backups", "/api/v1/restores", "/api/v1/restore-verifications"];
const wsPaths = ["/ws/v1/backups", "/ws/v1/restores", "/ws/v1/restore-verifications"];
function request(path, headers = {}) {
  return new Promise((resolve, reject) => {
    const req = http.request(new URL(path, base), { method: "GET", headers }, (res) => {
      res.resume();
      res.on("end", () => resolve(res.statusCode));
    });
    req.on("error", reject);
    req.end();
  });
}
(async () => {
  const results = [];
  for (const path of httpPaths) {
    results.push({ path, status: await request(path) });
  }
  for (const path of wsPaths) {
    results.push({
      path,
      status: await request(path, {
        Connection: "Upgrade",
        Upgrade: "websocket",
        "Sec-WebSocket-Key": "dGhlIHNhbXBsZSBub25jZQ==",
        "Sec-WebSocket-Version": "13",
        Origin: base.origin,
      }),
    });
  }
  fs.writeFileSync(output, `${JSON.stringify({ schema_id: "cartulary.standup_public_recovery_route_absence.v1", results })}\n`);
  for (const result of results) {
    if (result.status !== 404) {
      throw new Error(`${result.path} returned ${result.status}, want 404`);
    }
  }
})().catch((error) => {
  console.error(error.message);
  process.exit(1);
});
EOF

"$NODE" - "$capture_json" "$latest_json" "$restore_verify_json" "$route_json" "$summary_json" <<'EOF'
const fs = require("node:fs");
const [capturePath, latestPath, duePath, routePath, summaryPath] = process.argv.slice(2);
const capture = JSON.parse(fs.readFileSync(capturePath, "utf8"));
const latest = JSON.parse(fs.readFileSync(latestPath, "utf8"));
const due = JSON.parse(fs.readFileSync(duePath, "utf8"));
const routes = JSON.parse(fs.readFileSync(routePath, "utf8"));
const summary = {
  schema_id: "cartulary.standup_operational_recovery_smoke.v1",
  result: "pass",
  backup_set_id: latest.backup_set_id,
  captured_backup_set_id: capture.backup_set_id,
  consistency_point_at: latest.consistency_point_at,
  retained_until: latest.retained_until,
  restore_verification_due_count: due.due_count,
  restore_verification_verified_count: due.verified_count,
  public_route_absence_count: routes.results.length,
};
fs.writeFileSync(summaryPath, `${JSON.stringify(summary, null, 2)}\n`);
EOF

echo "standup-operational-recovery-smoke verified: backup capture, latest metadata, retention, due restore verification, public route absence, and retained artifacts."
