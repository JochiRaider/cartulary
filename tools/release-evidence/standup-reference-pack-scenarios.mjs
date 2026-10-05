// Exercises shipped entry points. SQL is test observation and explicit loss
// injection only; no alternate admission, verification, or recovery executor.
import assert from "node:assert/strict";
import { createHash, createHmac, randomUUID } from "node:crypto";
import { chmodSync, copyFileSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { setTimeout as delay } from "node:timers/promises";

const [root, work, project, origin, artifacts] = process.argv.slice(2);
assert.match(project, /^cartularymvpsmoke[0-9]+$/);
assert.match(path.basename(work), /^cartulary-standup-package-smoke\.[a-zA-Z0-9]+$/);
const checks = [];
function completed(name) {
  checks.push(name);
  writeFileSync(path.join(artifacts, "scenarios.json"), JSON.stringify({ completed: checks }, null, 2));
  process.stdout.write(`package scenario passed: ${name}\n`);
}
function command(binary, args, options = {}) {
  const result = spawnSync(binary, args, { encoding: "utf8", timeout: 180000, maxBuffer: 16 * 1024 * 1024, ...options });
  if (result.status !== 0) {
    let reason = "";
    try { const value = JSON.parse(result.stdout); reason = `${value.error?.code ?? ""}/${value.error?.details?.reason_code ?? value.error?.reason_code ?? ""}`; } catch { /* never retain raw command output */ }
    throw new Error(`package command failed: ${path.basename(binary)} ${args[0] ?? ""} (status ${result.status}; ${reason})`);
  }
  return result.stdout.trim();
}
function deployment(directory, name, url) {
  const composeArgs = ["compose", "--project-name", name, "--env-file", path.join(directory, ".env"), "-f", path.join(directory, "docker-compose.yml"), "-f", path.join(directory, "reference-packs.yml")];
  const compose = (...args) => command("docker", [...composeArgs, ...args]);
  const sql = (query, database = "cartulary") => compose("exec", "-T", "postgres", "psql", "-v", "ON_ERROR_STOP=1", "-U", "cartulary", "-d", database, "-Atc", query);
  const cookies = new Map();
  async function request(route, body, { status = 200, bearer, upload } = {}) {
    const headers = {};
    if (bearer) headers.Authorization = `Bearer ${bearer}`;
    else if (cookies.size) {
      headers.Cookie = [...cookies].map(([key, value]) => `${key}=${value}`).join("; ");
      headers["X-CSRF-Token"] = cookies.get("cartulary_csrf");
    }
    let payload;
    if (upload) {
      const boundary = "cartulary-package-smoke-boundary";
      headers["Content-Type"] = `multipart/form-data; boundary=${boundary}`;
      payload = Buffer.concat([Buffer.from(`--${boundary}\r\nContent-Disposition: form-data; name="metadata"\r\nContent-Type: application/json\r\n\r\n${JSON.stringify(body)}\r\n--${boundary}\r\nContent-Disposition: form-data; name="file"; filename="bundle.${upload.zip ? "zip" : "tar"}"\r\nContent-Type: ${upload.zip ? "application/zip" : "application/x-tar"}\r\n\r\n`), upload.bytes, Buffer.from(`\r\n--${boundary}--\r\n`)]);
    } else if (body !== undefined) {
      headers["Content-Type"] = "application/json";
      payload = JSON.stringify(body);
    }
    const response = await fetch(url + route, { method: body === undefined ? "GET" : "POST", headers, body: payload, signal: AbortSignal.timeout(30000) });
    const value = await response.json();
    for (const cookie of response.headers.getSetCookie()) {
      const [key, ...rest] = cookie.split(";", 1)[0].split("=");
      cookies.set(key, rest.join("="));
    }
    assert.equal(response.status, status, `${route}: ${value.error?.code ?? "unexpected status"}/${value.error?.details?.reason_code ?? ""}`);
    return status < 400 ? value.data : value.error;
  }
  async function ready(want = 200) {
    for (let i = 0; i < 120; i++) {
      try { if ((await fetch(url + "/readyz", { signal: AbortSignal.timeout(2000) })).status === want) return; } catch { /* bounded startup wait */ }
      await delay(1000);
    }
    throw new Error(`package readiness did not reach ${want}`);
  }
  async function job(admitted, want = "succeeded") {
    for (let i = 0; i < 120; i++) {
      const value = await request(`/api/v1/jobs/${admitted.job_id}`);
      if (["succeeded", "failed", "canceled"].includes(value.status)) {
        assert.equal(value.status, want, `Job: ${value.error_summary?.code}/${value.error_summary?.details?.reason_code}`);
        return value;
      }
      await delay(500);
    }
    throw new Error("package Job did not finish");
  }
  const post = (route, txn, extra = {}, options) => request(route, { client_txn_id: txn, ...extra }, options);
  const action = (version, operation, options) => post(`/api/v1/reference-packs/enrichment.tor/${version}/${operation}`, `${operation}-${version}`, { reason: "Disposable package qualification" }, options);
  return { directory, name, url, compose, sql, request, ready, job, post, action };
}
function totp(secret) {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
  const bits = [...secret.replace(/=+$/, "")].map((letter) => alphabet.indexOf(letter).toString(2).padStart(5, "0")).join("");
  const bytes = Buffer.from(bits.match(/.{8}/g).map((byte) => Number.parseInt(byte, 2)));
  const counter = Buffer.alloc(8); counter.writeBigUInt64BE(BigInt(Math.floor(Date.now() / 30000)));
  const hash = createHmac("sha1", bytes).update(counter).digest();
  return String((hash.readUInt32BE(hash[19] & 15) & 0x7fffffff) % 1000000).padStart(6, "0");
}
async function login(stack) {
  const credentials = { username: "admin@example.test", password: "ReplaceThisBootstrap1!" };
  const rejection = await stack.request("/api/v1/auth/login", credentials, { status: 401 });
  assert.equal(rejection.code, "mfa_setup_required");
  const bearer = rejection.details.bootstrap_token;
  const setup = await stack.post("/api/v1/auth/mfa/totp/begin", "enroll-begin", {}, { bearer });
  await stack.post("/api/v1/auth/mfa/totp/complete", "enroll-complete", { enrollment_id: setup.enrollment_id, code: totp(setup.totp_setup.secret_base32) }, { bearer });
  await stack.request("/api/v1/auth/login", { ...credentials, second_factor: { kind: "totp", assertion: { code: totp(setup.totp_setup.secret_base32) } } });
  return (await stack.request("/api/v1/auth/session")).user_id;
}
const source = deployment(work, project, origin);
const selection = (stack) => stack.sql("SELECT pack_set_id FROM reference_pack_current_set WHERE singleton");
const published = (stack) => stack.sql("SELECT json_build_array((SELECT count(*) FROM reference_pack_versions),(SELECT count(*) FROM reference_pack_envelopes),(SELECT count(*) FROM reference_pack_indexes),(SELECT count(*) FROM reference_pack_roots),(SELECT md5(string_agg(t::text,',' ORDER BY repository_id,pack_key,pack_version,role)) FROM reference_pack_metadata_versions t),(SELECT md5(string_agg(t::text,',' ORDER BY repository_id)) FROM reference_pack_repositories t),(SELECT pack_set_id FROM reference_pack_current_set WHERE singleton))");
const durable = (stack) => stack.sql("SELECT json_build_array((SELECT count(*) FROM reference_pack_candidates),(SELECT count(*) FROM reference_pack_operations),(SELECT count(*) FROM jobs),(SELECT count(*) FROM reference_pack_events))");
const file = (name) => readFileSync(path.join(work, "reference-pack-incoming", name));
const upload = (stack, name, txn, status = 202) => stack.post("/api/v1/reference-packs/import", txn, {}, { status, upload: { bytes: file(name) } });
const id = (value) => { assert.match(value, /^[a-zA-Z0-9_.-]+$/); return value; };

// Enable the consumer owners in this disposable deployment explicitly.
source.compose("stop", "app");
writeFileSync(path.join(work, "config.toml"), readFileSync(path.join(work, "config.toml"), "utf8") + "\n[snapshot_reporting]\nclaimed = true\n[incident_portability]\nclaimed = true\n");
source.compose("up", "-d", "--no-deps", "--force-recreate", "app");
await source.ready();
const actor = id(await login(source));
assert.equal(source.sql("SELECT count(*) FROM reference_pack_operations o JOIN jobs j USING(job_id) WHERE o.kind='import' AND o.actor_kind='local_operator' AND o.actor_user_id IS NULL AND o.terminal_at IS NOT NULL AND j.status='succeeded'"), "2");
assert.equal(source.sql("SELECT count(*) FROM reference_pack_current_set s JOIN reference_pack_set_members m USING(pack_set_id) WHERE m.pack_key='enrichment.tor'"), "0");
completed("operator actor and staged-only import");

const beforeMalformed = durable(source);
await source.post("/api/v1/reference-packs/import", "invalid-metadata", { activation_policy: "activate" }, { status: 400, upload: { bytes: file("valid.tar") } });
assert.equal(durable(source), beforeMalformed, "pre-admission rejection mutated durable state");
completed("invalid transport metadata has no durable effects");
const beforeMalformedPublication = published(source);
const malformedJob = await upload(source, "malformed.tar", "malformed");
await source.job(malformedJob, "failed");
assert.equal(published(source), beforeMalformedPublication);
assert.equal(source.sql(`SELECT count(*) FROM reference_pack_operations WHERE job_id='${id(malformedJob.job_id)}' AND terminal_at IS NOT NULL AND final_outcome IS NOT NULL`), "1");
completed("malformed archive retains terminal failure without publication");

await source.action(2, "activate");
const selected = selection(source);
const incident = await source.post("/api/v1/incidents", "incident", { incident_key: "PACK-QUALIFICATION", title: "Package reference history" }, { status: 201 });
const incidentID = id(incident.incident_id);
const snapshotJob = await source.post("/api/v1/snapshots", "snapshot", { incident_id: incidentID }, { status: 202 });
const snapshot = id((await source.job(snapshotJob)).result_summary.resource_refs.find((ref) => ref.type === "snapshot" || ref.kind === "snapshot").id);
const bindingSQL = `SELECT export_model_json->'reference_packs' FROM reporting_snapshots WHERE snapshot_id='${snapshot}'`;
const binding = source.sql(bindingSQL);
assert.equal(JSON.parse(binding).pack_set_id, selected);
completed("snapshot pins exact activated set");

const rotatedJob = await upload(source, "rotated.tar", "rotate-import");
await source.job(rotatedJob);
assert.equal((await upload(source, "rotated.tar", "rotate-import")).job_id, rotatedJob.job_id);
assert.equal(source.sql(`SELECT actor_kind||':'||actor_user_id FROM reference_pack_operations WHERE job_id='${id(rotatedJob.job_id)}'`), `user:${actor}`);
assert.equal(source.sql("SELECT root_version FROM reference_pack_repositories WHERE repository_id='package_smoke.repo'"), "2");
assert.equal(selection(source), selected, "import implicitly activated");
await source.action(3, "activate");
assert.notEqual(selection(source), selected);
const activeRemoval = await source.action(3, "remove", { status: 409 });
assert.equal(activeRemoval.details.reason_code, "active");
source.compose("restart", "app"); await source.ready();
assert.equal(source.sql("SELECT root_version FROM reference_pack_repositories WHERE repository_id='package_smoke.repo'"), "2");
completed("HTTP actor, idempotent replay, explicit activation, root rotation and restart");

for (const [name, reason] of [["rollback.tar", "metadata_rollback_detected"], ["bad-signature.tar", "signature_threshold_not_met"], ["incompatible.tar", "contract_incompatible"]]) {
  const state = published(source);
  const admitted = await upload(source, name, name);
  const terminal = await source.job(admitted, "failed");
  assert.equal(terminal.error_summary?.details?.reason_code, reason);
  assert.equal(published(source), state, `${name}: failed verification partially published`);
  assert.equal(source.sql(`SELECT count(*) FROM reference_pack_operations WHERE job_id='${id(admitted.job_id)}' AND terminal_at IS NOT NULL AND final_outcome IS NOT NULL`), "1");
  assert.equal(source.sql(`SELECT count(*) FROM reference_pack_events e JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id='${id(admitted.job_id)}'`), "1");
}
completed("admitted failures retain terminal audit without partial publication");

for (const txn of ["render", "rerender"]) {
  const admitted = await source.post("/api/v1/releases", txn, { snapshot_id: snapshot, template_id: "cartulary.report.default", template_version: "1", redaction_profile_id: "cartulary.redaction.internal", redaction_profile_version: "1", release_scope: "internal_draft", output_kind: "slidev" }, { status: 202 });
  const terminal = await source.job(admitted);
  const release = id(terminal.result_summary.resource_refs.find((ref) => ref.type === "release" || ref.kind === "release").id);
  assert.deepEqual(JSON.parse(source.sql(`SELECT bundle_manifest_json->'reference_packs' FROM reporting_render_bundles WHERE release_id='${release}'`)), JSON.parse(binding));
}
const pinned = await source.action(2, "remove", { status: 409 });
assert.equal(pinned.details.reason_code, "pinned");
completed("historical report rerender and pin-protected removal");

// A distinct disposable Compose project is the portability destination. It has
// its own database, bucket, volumes, trust bootstrap and bootstrap administrator.
const destinationWork = path.join(work, "destination");
mkdirSync(destinationWork, { mode: 0o700 });
for (const name of ["docker-compose.yml", "reference-packs.yml", "config.toml", "bootstrap-admin.json", "revisions-conflict-token-key-ring.json", "postgres-provision.sh", "reference-pack-trust.json"]) {
  copyFileSync(path.join(work, name), path.join(destinationWork, name));
}
mkdirSync(path.join(destinationWork, "reference-pack-incoming"), { mode: 0o755 });
const portServer = (await import("node:net")).createServer();
await new Promise((resolve) => portServer.listen(0, "127.0.0.1", resolve));
const destinationPort = portServer.address().port;
await new Promise((resolve) => portServer.close(resolve));
const destinationOrigin = `http://127.0.0.1:${destinationPort}`;
writeFileSync(path.join(destinationWork, ".env"), readFileSync(path.join(work, ".env"), "utf8").replace(/^CARTULARY_HTTP_PORT=.*$/m, `CARTULARY_HTTP_PORT=${destinationPort}`).replace(/^CARTULARY_PUBLIC_ORIGIN=.*$/m, `CARTULARY_PUBLIC_ORIGIN=${destinationOrigin}`));
const destination = deployment(destinationWork, `${project}destination`, destinationOrigin);
destination.compose("up", "-d", "app"); await destination.ready(); await login(destination);
const destinationBase = selection(destination);
const exportJob = await source.post("/api/v1/incident-bundles/export", "export", { incident_id: incidentID, optional_sections: ["snapshots"], reference_pack_mode: "embedded" }, { status: 202 });
const exported = await source.job(exportJob);
const bundleID = id(exported.result_summary.resource_refs[0].id);
const bundleRef = source.sql(`SELECT bundle_storage_ref FROM incident_bundle_exports WHERE bundle_id='${bundleID}'`);
assert.match(bundleRef, /^[a-zA-Z0-9_./-]+$/); assert.ok(!bundleRef.startsWith("/") && !bundleRef.split("/").includes(".."));
source.compose("cp", `app:/var/lib/cartulary/exports/${bundleRef}`, path.join(work, "portable.zip"));
const imported = await destination.post("/api/v1/incident-bundles/import", "import", {}, { status: 202, upload: { zip: true, bytes: readFileSync(path.join(work, "portable.zip")) } });
await destination.job(imported);
assert.equal(selection(destination), destinationBase, "portability transported source activation");
assert.equal(destination.sql("SELECT count(*) FROM reporting_snapshots"), "0", "portability fabricated native snapshots");
const portableCatalog = JSON.parse(destination.sql("SELECT convert_from(content,'UTF8') FROM reporting_imported_artifact_files WHERE bundle_path='ext/snapshots/catalog.json'"));
assert.equal(portableCatalog.snapshots.length, 1);
const portableModelPath = portableCatalog.snapshots[0].export_model_path;
assert.match(portableModelPath, /^[a-zA-Z0-9_./-]+$/);
assert.deepEqual(JSON.parse(destination.sql(`SELECT convert_from(content,'UTF8')::json->'reference_packs' FROM reporting_imported_artifact_files WHERE bundle_path='${portableModelPath}'`)), JSON.parse(binding));
assert.equal(destination.sql("SELECT count(*) FROM reference_pack_candidates WHERE pack_key='enrichment.tor' AND pack_version='2' AND health='verified_available'"), "1");
completed("destination-trusted portability preserves history without activation");

// Use the shipped backup and restore-verification wrappers, preserving overlay.
copyFileSync(path.join(root, "deploy/mvp/restore-verification-target.toml.example"), path.join(work, "restore-verification-target.toml"));
chmodSync(path.join(work, "restore-verification-target.toml"), 0o644);
const restoreRoot = path.join(work, "runtime/restore-verification-target");
mkdirSync(restoreRoot, { recursive: true }); chmodSync(restoreRoot, 0o777);
const recoveryEnv = { ...process.env, CARTULARY_MVP_DIR: work, CARTULARY_MVP_COMPOSE_PROJECT_NAME: project, CARTULARY_MVP_COMPOSE_OVERLAY: path.join(work, "reference-packs.yml") };
const backup = JSON.parse(command(path.join(root, "deploy/mvp/scripts/backup-capture.sh"), [], { env: recoveryEnv }));
assert.equal(backup.result, "succeeded"); await source.ready();
const verified = JSON.parse(command(path.join(root, "deploy/mvp/scripts/restore-verify-due.sh"), [], { env: recoveryEnv }));
assert.equal(verified.result, "succeeded"); assert.equal(verified.backup_set_id, backup.backup_set_id);
assert.equal(source.sql("SELECT count(*) FROM reporting_snapshots", "cartulary_restore_verify"), "0", "verification did not reset its disposable target");
// Due verification deliberately resets its target. Admit that empty target for
// an actual restore with a fresh generation and the explicit restore purpose.
const markerPath = path.join(restoreRoot, "backups/restore-target-marker.json");
const marker = JSON.parse(readFileSync(markerPath, "utf8"));
marker.purpose = "restore_target";
marker.target_generation_id = randomUUID();
writeFileSync(path.join(restoreRoot, "backups/restore-target-generation"), marker.target_generation_id + "\n");
writeFileSync(markerPath, JSON.stringify(marker) + "\n");
const restored = JSON.parse(source.compose("run", "--rm", "--no-deps",
  "--volume", `${path.join(work, "config.toml")}:/etc/cartulary/config.toml:ro`,
  "--volume", `${path.join(work, "restore-verification-target.toml")}:/etc/cartulary/restore-verification-target.toml:ro`,
  "--volume", `${restoreRoot}:/var/lib/cartulary/restore-verification-target`,
  "--entrypoint", "/usr/local/bin/cartulary-operator", "restore-verify-operator", "restore", "latest",
  "--source-config-file", "/etc/cartulary/config.toml",
  "--target-config-file", "/etc/cartulary/restore-verification-target.toml",
  "--confirm-backup-set-id", backup.backup_set_id));
assert.equal(restored.result, "succeeded"); assert.equal(restored.backup_set_id, backup.backup_set_id);
assert.deepEqual(JSON.parse(source.sql(bindingSQL, "cartulary_restore_verify")), JSON.parse(binding));
assert.equal(source.sql("SELECT root_version FROM reference_pack_repositories WHERE repository_id='package_smoke.repo'", "cartulary_restore_verify"), "2");
const recoveryHelperImage = command("docker", ["inspect", "--format", "{{.Config.Image}}", source.compose("ps", "-q", "postgres")]);
const restoredExportHash = command("docker", ["run", "--rm", "--label", `com.docker.compose.project=${project}`, "--entrypoint", "sha256sum", "--mount", `type=bind,source=${restoreRoot},target=/target,readonly`, recoveryHelperImage, `/target/exports/${bundleRef}`]).split(" ")[0];
assert.equal(restoredExportHash, createHash("sha256").update(readFileSync(path.join(work,"portable.zip"))).digest("hex"), "restore changed exported Incident Bundle bytes");
writeFileSync(path.join(artifacts, "recovery.json"), JSON.stringify({ backup_set_id: backup.backup_set_id, restore_result: restored.result, snapshot_id: snapshot, pack_set_id: selected }));
completed("matching-package restore preserves historical binding and rotated trust");

// This version is created after backup and never pinned by a consumer.
await source.job(await upload(source, "collectable.tar", "collectable"));
const refs = source.sql("SELECT o.storage_ref FROM reference_pack_objects o JOIN reference_pack_object_refs r USING(object_id) JOIN reference_pack_versions v ON o.sha256=v.manifest_sha256 WHERE r.owner_kind='version' AND v.pack_key='enrichment.tor' AND v.pack_version='7'").split("\n");
assert.ok(refs.length > 0 && refs[0]);
await source.action(7, "remove"); source.compose("restart", "app"); await source.ready();
for (const ref of refs) {
  assert.match(ref, /^[a-zA-Z0-9_./-]+$/); assert.ok(!ref.startsWith("/") && !ref.split("/").includes(".."));
  assert.equal(source.sql(`SELECT available FROM reference_pack_objects WHERE storage_ref='${ref}'`), "f");
  const result = spawnSync("docker", ["cp", `${source.compose("ps", "-q", "app")}:/var/lib/cartulary/reference-packs/${ref}`, path.join(work, "unexpected-object")], { encoding: "utf8", timeout: 30000 });
  assert.notEqual(result.status, 0, "unreferenced extracted object survived collection");
}
completed("unpinned removal and physical collection");

// Inject definitive loss into this test allocation only, then exercise the
// consumer boundary and readiness. No retained deployment is selectable.
const lostRef = source.sql("SELECT o.storage_ref FROM reference_pack_objects o JOIN reference_pack_versions v ON o.sha256=v.manifest_sha256 WHERE v.pack_key='type_registry.host' AND o.available LIMIT 1");
assert.match(lostRef, /^[a-zA-Z0-9_./-]+$/); assert.ok(!lostRef.startsWith("/") && !lostRef.split("/").includes(".."));
const helperImage = command("docker", ["inspect", "--format", "{{.Config.Image}}", source.compose("ps", "-q", "postgres")]);
command("docker", ["run", "--rm", "--label", `com.docker.compose.project=${project}`, "--entrypoint", "sh", "-v", `${project}_cartulary-reference-packs:/data`, helperImage, "-c", 'rm -- "/data/$1"', "sh", lostRef]);
await source.post("/api/v1/snapshots", "loss", { incident_id: incidentID }, { status: 409 });
await source.ready(503);
assert.equal(source.sql("SELECT count(*) FROM reference_pack_current_set WHERE pack_set_id IS NULL"), "1");
completed("required-content loss fails consumer admission and readiness");
