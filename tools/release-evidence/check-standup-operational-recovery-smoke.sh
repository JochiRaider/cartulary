#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../.." && pwd)"
PACKAGE_DIR="$ROOT_DIR/deploy/mvp"
NODE="${NODE_BIN:-node}"

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
port="$(pick_port)" || fail "no free loopback port found for operational recovery smoke"
public_origin="https://127.0.0.1:${port}"
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

# shellcheck source=tools/release-evidence/package-smoke-cleanup.sh
source "$ROOT_DIR/tools/release-evidence/package-smoke-cleanup.sh"
cleanup() {
  local status=$?
  trap - EXIT INT TERM
  set +e
  local artifact
  for artifact in "$capture_json" "$latest_json" "$restore_verify_json" "$route_json" "$summary_json"; do
    if [[ -f "$artifact" ]] && ! cp "$artifact" "$artifact_dir/"; then status=1; fi
  done
  local helper_image="${package_postgres_image:-}"
  if [[ -z "$helper_image" ]] && rmdir -- "$work_dir"; then
    printf '{"cleanup":"passed","empty_preinstallation_workspace":true}\n' >"$artifact_dir/workspace-cleanup.json"
  elif [[ -z "$helper_image" ]] || ! cleanup_package_workspace "$project" "$work_dir" "$helper_image" "$artifact_dir/workspace-cleanup.json"; then status=1; fi
  if ! cleanup_package_resources "${project}destination" "" "$artifact_dir/destination-cleanup.json"; then status=1; fi
  if ! cleanup_package_resources "$project" "" "$artifact_dir/cleanup.json"; then status=1; fi
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# shellcheck source=tools/release-evidence/consume-package.sh
source "$ROOT_DIR/tools/release-evidence/consume-package.sh"
consume_package_release "$ROOT_DIR" "$work_dir" "$artifact_dir"
compose_file="$PACKAGE_DIR/docker-compose.yml"
cp "$PACKAGE_DIR/config.toml.example" "$work_dir/config.toml"
# shellcheck source=tools/release-evidence/package-fixture-tls.sh
source "$ROOT_DIR/tools/release-evidence/package-fixture-tls.sh"
provision_package_fixture_tls "$ROOT_DIR" "$PACKAGE_DIR" "$work_dir"
cp "$PACKAGE_DIR/bootstrap-admin.json.example" "$work_dir/bootstrap-admin.json"
cp "$PACKAGE_DIR/revisions-conflict-token-key-ring.json.example" "$work_dir/revisions-conflict-token-key-ring.json"
cp "$PACKAGE_DIR/restore-verification-target.toml.example" "$work_dir/restore-verification-target.toml"
chmod 0644 \
  "$work_dir/config.toml" \
  "$work_dir/bootstrap-admin.json" \
  "$work_dir/revisions-conflict-token-key-ring.json" \
  "$work_dir/restore-verification-target.toml"

cat >"$work_dir/.env" <<EOF
CARTULARY_MVP_COMPOSE_PROJECT_NAME=${project}
CARTULARY_HTTP_PORT=${port}
CARTULARY_PUBLIC_ORIGIN=${public_origin}

POSTGRES_DB=cartulary
CARTULARY_TLS_DIR=${work_dir}/tls

CARTULARY_AUTH_MASTER_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=
CARTULARY_RECOVERY_MASTER_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=
CARTULARY_SECRET_REVISIONS_CONFLICT_TOKEN_ACTIVE=cmV2aXNpb25zLXRva2VuLWtleS1tYXRlcmlhbC0wMDE

CARTULARY_S3_PRIMARY_ACCESS_KEY_ID=cartulary-local
CARTULARY_S3_PRIMARY_SECRET_ACCESS_KEY=cartulary-local-secret
CARTULARY_S3_PRIMARY_BUCKET=cartulary-mvp-smoke

RESTORE_VERIFY_POSTGRES_DB=cartulary_restore_verify
CARTULARY_S3_RESTORE_VERIFY_ENDPOINT=seaweedfs-s3:8333
CARTULARY_S3_RESTORE_VERIFY_ACCESS_KEY_ID=cartulary-local
CARTULARY_S3_RESTORE_VERIFY_SECRET_ACCESS_KEY=cartulary-local-secret
CARTULARY_S3_RESTORE_VERIFY_BUCKET=cartulary-mvp-restore-verify-smoke
EOF

"${NODE_BIN:-node}" "$ROOT_DIR/tools/release-evidence/package-platform.mjs" prepare "$work_dir" "$artifact_dir" "$project" "$public_origin"

CARTULARY_MVP_DIR="$work_dir" CARTULARY_MVP_ENV_FILE="$work_dir/.env" \
  "$PACKAGE_DIR/scripts/package.sh" start >"$work_dir/startup.log" || {
  compose logs --no-color --tail 80 postgres seaweedfs-s3 migrate object-store-init >&2 || true
  fail "package initialization failed"
}

wait_for_http_status() {
  local path="$1"
  local want="$2"
  local output="$3"
  local start_time="$SECONDS"
  local status="000"
  while ((SECONDS - start_time < 180)); do
    status="$(curl --cacert "$work_dir/tls/ca.pem" --tlsv1.3 -sS -o "$output" -w '%{http_code}' "${public_origin}${path}" || true)"
    if [[ "$status" == "$want" ]]; then
      return 0
    fi
    sleep 1
  done
  echo "last status for ${path}: ${status}" >&2
  return 1
}

wait_for_http_status "/readyz" "200" "$ready_body" || fail "/readyz did not become ready"
"${NODE_BIN:-node}" "$ROOT_DIR/tools/release-evidence/package-platform.mjs" observe "$work_dir" "$artifact_dir" "$project" "$public_origin"

grep -Fq '"status":"ready"' "$ready_body" || fail "/readyz did not report structured ready status"
"${NODE_BIN:-node}" "$ROOT_DIR/tools/release-evidence/recovery-interruption.mjs" "$work_dir" "$project" "$artifact_dir"
wait_for_http_status "/readyz" "200" "$ready_body" || fail "interrupted package did not become ready"


CARTULARY_MVP_DIR="$work_dir" \
  CARTULARY_MVP_ENV_FILE="$work_dir/.env" \
  "$PACKAGE_DIR/scripts/package.sh" backup-create >"$capture_json"

compose run --rm --no-deps \
  --entrypoint /usr/local/bin/cartulary-operator \
  recovery-operator backup inspect latest \
  --source-config-file /etc/cartulary/config.toml >"$latest_json"

"$NODE" - "$capture_json" "$latest_json" <<'EOF'
const fs = require("node:fs");
const capture = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const latest = JSON.parse(fs.readFileSync(process.argv[3], "utf8"));
const dayMs = 24 * 60 * 60 * 1000;
function fail(message) {
  console.error(message);
  process.exit(1);
}
function requireRecoveryResult(payload, operation) {
  if (payload.schema_id !== "cartulary.operator_recovery_result.v2") {
    fail(`unexpected ${operation} schema_id ${payload.schema_id}`);
  }
  if (payload.operation !== operation || payload.result !== "succeeded" || payload.error !== null) {
    fail(`unexpected ${operation} result ${JSON.stringify(payload)}`);
  }
  if (typeof payload.operation_id !== "string" || payload.operation_id.length === 0) {
    fail(`${operation} missing operation_id`);
  }
  if (typeof payload.backup_set_id !== "string" || payload.backup_set_id.length === 0) {
    fail(`${operation} missing backup_set_id`);
  }
  if (!Array.isArray(payload.artifact_refs) || payload.artifact_refs.length === 0) {
    fail(`${operation} missing artifact_refs`);
  }
}
requireRecoveryResult(capture, "backup_create");
requireRecoveryResult(latest, "backup_inspect_latest");
if (capture.backup_set_id !== latest.backup_set_id) {
  fail("latest metadata did not select captured backup_set");
}
const consistency = new Date(latest.consistency_point_at).getTime();
if (!Number.isFinite(consistency) || Date.now() - consistency > dayMs) {
  fail("latest backup consistency point is older than 24 hours");
}
for (const ref of latest.artifact_refs) {
  if (typeof ref.kind !== "string" || typeof ref.schema_id !== "string" || typeof ref.ref_id !== "string") {
    fail(`latest backup contains malformed artifact ref ${JSON.stringify(ref)}`);
  }
  for (const forbidden of ["/", "\\", "postgresql://", "seaweedfs-s3", "cartulary-mvp-smoke"]) {
    if (ref.ref_id.includes(forbidden)) {
      fail(`latest backup artifact ref exposes unsafe detail ${JSON.stringify(ref)}`);
    }
  }
}
EOF

CARTULARY_MVP_DIR="$work_dir" \
  CARTULARY_MVP_ENV_FILE="$work_dir/.env" \
  "$PACKAGE_DIR/scripts/package.sh" restore-verify-due >"$restore_verify_json"

# Repeated no-due maintenance must preserve the owner-issued proof exactly.
proof_digest() {
  compose run --rm --no-deps --user 0 \
    --volume "$work_dir/runtime/restore-verification-target:/target:ro" \
    --entrypoint /bin/sh postgres -c 'sha256sum /target/backups/restore-target-marker.json /target/backups/restore-target-generation'
}
proof_before="$(proof_digest)"
CARTULARY_MVP_DIR="$work_dir" CARTULARY_MVP_ENV_FILE="$work_dir/.env" \
  "$PACKAGE_DIR/scripts/package.sh" restore-verify-due >"$work_dir/repeated-verification.json"
[[ "$proof_before" == "$(proof_digest)" ]] || fail "no-op verification changed target proof"
"$NODE" - "$work_dir/repeated-verification.json" <<'EOF'
const fs = require("node:fs");
const result = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
if (result.result !== "no_op") throw new Error("repeated due verification was not a no-op");
EOF

"$NODE" - "$restore_verify_json" <<'EOF'
const fs = require("node:fs");
const due = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
function fail(message) {
  console.error(message);
  process.exit(1);
}
if (due.schema_id !== "cartulary.operator_recovery_result.v2") {
  fail(`unexpected due schema_id ${due.schema_id}`);
}
if (due.operation !== "restore_verify_due" || due.result !== "succeeded" || due.error !== null) {
  fail(`unexpected due result ${JSON.stringify(due)}`);
}
if (typeof due.backup_set_id !== "string" || due.backup_set_id.length === 0) {
  fail("due restore verification missing backup_set_id");
}
if (!Array.isArray(due.artifact_refs) || !due.artifact_refs.some((ref) => ref.kind === "restore_verification")) {
  fail(`due restore verification missing restore_verification artifact ref ${JSON.stringify(due)}`);
}
for (const ref of due.artifact_refs) {
  if (typeof ref.ref_id !== "string" || ref.ref_id.includes("/") || ref.ref_id.includes("\\")) {
    fail(`due restore verification contains unsafe artifact ref ${JSON.stringify(ref)}`);
  }
}
EOF

"$NODE" - "$public_origin" "$route_json" "$work_dir/tls/ca.pem" <<'EOF'
const fs = require("node:fs");
const https = require("node:https");
const ca = fs.readFileSync(process.argv[4]);
const base = new URL(process.argv[2]);
const output = process.argv[3];
const httpPaths = ["/api/v1/backups", "/api/v1/restores", "/api/v1/restore-verifications"];
const wsPaths = ["/ws/v1/backups", "/ws/v1/restores", "/ws/v1/restore-verifications"];
function request(path, headers = {}) {
  return new Promise((resolve, reject) => {
    const req = https.request(new URL(path, base), { method: "GET", headers, ca, minVersion: "TLSv1.3", maxVersion: "TLSv1.3", timeout: 5000 }, (res) => {
      res.resume();
      res.on("end", () => resolve(res.statusCode));
    });
    req.on("timeout", () => req.destroy(new Error("package HTTPS probe timed out")));
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

"$NODE" "$ROOT_DIR/tools/release-evidence/portable-package-recovery.mjs" "$work_dir" "$project" "$public_origin" "$artifact_dir"

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
  latest_operation_id: latest.operation_id,
  restore_verification_result: due.result,
  restore_verification_artifact_count: due.artifact_refs.length,
  public_route_absence_count: routes.results.length,
};
fs.writeFileSync(summaryPath, `${JSON.stringify(summary, null, 2)}\n`);
EOF

echo "standup-operational-recovery-smoke verified: backup create, latest inspect, due restore verification, public route absence, and retained artifacts."
