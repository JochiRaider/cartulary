// The browser stack's private records travel on standard input, so secrets
// and transport values never enter a process environment (P6-4).
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, statSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

const root = path.resolve(import.meta.dirname, "../../../..");
const cli = path.join(root, "tools/harness/browser/web-e2e-session-cli.mjs");
const manifest = path.join(root, "tools/harness_redaction_manifest.json");
const marker = JSON.parse(readFileSync(manifest, "utf8")).replacement;
const run = (args, records) => spawnSync(process.execPath, [cli, ...args], { input: records.map((record) => `${record}\0`).join(""), env: { PATH: process.env.PATH }, encoding: "utf8" });

function scratch(t) {
  const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-web-e2e-session-"));
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  return directory;
}

test("restart secrets are written privately from records", (t) => {
  const file = path.join(scratch(t), "private", "secrets.json");
  const result = run(["write-restart-secrets", file], ["dsn=postgres://user:pw@host/db", "auth_master_key=key with = sign", "s3_secret_access_key=secret"]);
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(JSON.parse(readFileSync(file, "utf8")), { dsn: "postgres://user:pw@host/db", auth_master_key: "key with = sign", s3_secret_access_key: "secret" });
  assert.equal(statSync(file).mode & 0o777, 0o600);
});

test("a startup failure redacts every given secret and credential URL, across lines", () => {
  const secret = "very-secret-token-value";
  const result = run(["failure-message", "backend", manifest], [`text=line one ${secret}\nconnect postgres://u:p@db:5432 failed`, `secret=${secret}`, "secret=short"]);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout.includes(secret), false);
  assert.equal(result.stdout.includes("u:p@"), false);
  assert.equal(result.stdout, `backend process failed to establish a session: line one ${marker} connect ${marker} failed\n`);
});

test("the session environment omits empty values and the lease records the stack", (t) => {
  const directory = scratch(t);
  const envFile = path.join(directory, "session.env.json"), leaseFile = path.join(directory, "lease.json");
  const result = run(["write-session", envFile, leaseFile], ["env.CARTULARY_WEB_E2E_API_ORIGIN=https://api", "env.CARTULARY_FIXTURE_ROW_ID=",
    "lease.backend_port=8443", "lease.keep_runtime_root=1", "lease.server_pgid=42"]);
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(JSON.parse(readFileSync(envFile, "utf8")), { CARTULARY_WEB_E2E_API_ORIGIN: "https://api" });
  const lease = JSON.parse(readFileSync(leaseFile, "utf8"));
  assert.deepEqual([lease.schema_id, lease.backend_port, lease.keep_runtime_root, lease.server_pgid, lease.backend_generation, lease.session_env_file],
    ["cartulary.web_e2e_session_lease.v1", 8443, true, "42", 1, envFile]);
  const update = run(["update-lease", leaseFile], ["server_pgid=43", "backend_generation=2", "backend_generation_head=head-2"]);
  assert.equal(update.status, 0, update.stderr);
  const updated = JSON.parse(readFileSync(leaseFile, "utf8"));
  assert.deepEqual([updated.server_pgid, updated.backend_generation, updated.backend_generation_head, updated.env.CARTULARY_WEB_E2E_BACKEND_GENERATION_HEAD], ["43", 2, "head-2", "head-2"]);
});

test("malformed records and usage are rejected", (t) => {
  const directory = scratch(t);
  assert.notEqual(run(["write-session", path.join(directory, "e"), path.join(directory, "l")], ["unknown.key=1"]).status, 0);
  assert.notEqual(run(["write-restart-secrets", path.join(directory, "s")], ["no-separator"]).status, 0);
  assert.equal(run(["write-session", path.join(directory, "e")], []).status, 2);
});
