#!/usr/bin/env node
// The browser stack's private records (TH-HARNESS Section 5.5). The stack
// passes values on standard input as NUL-separated key=value records, never in
// an environment, so its secrets never appear in a process environment and no
// transport-only variable names exist.
//
//   write-restart-secrets <file>   write the backend restart secrets (mode 0600)
//   failure-message <label> <redaction-manifest>
//                                  print a bounded, redacted startup failure;
//                                  records: text, and secret for each value to
//                                  redact
//   write-session <env-file> <lease-file>
//                                  write the session environment and lease;
//                                  records: env.NAME for the session
//                                  environment, lease.FIELD for the lease
//   update-lease <lease-file>      record a restarted backend's process group
//                                  and generation
import fs from "node:fs";
import path from "node:path";

// records reads standard input as key=value records. A repeated key keeps
// every value, in order.
function records() {
  const input = fs.readFileSync(0, "utf8");
  const values = new Map();
  for (const record of input.split("\0")) {
    if (record === "") continue;
    const separator = record.indexOf("=");
    if (separator <= 0) throw new Error("malformed record");
    const key = record.slice(0, separator);
    values.set(key, [...(values.get(key) ?? []), record.slice(separator + 1)]);
  }
  return values;
}

const one = (values, key) => values.get(key)?.at(-1);

// writePrivate replaces file with a mode-0600 document.
function writePrivate(file, document) {
  fs.mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  const temporary = `${file}.tmp-${process.pid}`;
  fs.writeFileSync(temporary, `${JSON.stringify(document, null, 2)}\n`, { mode: 0o600, flag: "wx" });
  fs.renameSync(temporary, file);
  fs.chmodSync(file, 0o600);
}

const restartSecretFields = ["dsn", "auth_master_key", "revisions_secret", "cursor_secret", "safe_digest_secret", "s3_access_key_id", "s3_secret_access_key"];

function writeRestartSecrets([file]) {
  const values = records();
  writePrivate(file, Object.fromEntries(restartSecretFields.map((field) => [field, one(values, field)])));
}

function failureMessage([label, manifestFile]) {
  const values = records();
  const secrets = (values.get("secret") ?? []).filter((value) => value.length >= 8);
  // The single retained marker and the credential-bearing URL schemes are
  // owned by the redaction manifest (TH-HARNESS-REQ-617).
  const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
  const marker = manifest.replacement;
  const schemes = manifest.retained_scan?.url_userinfo_schemes;
  if (typeof marker !== "string" || marker === "") throw new Error("redaction manifest has no marker");
  if (!Array.isArray(schemes) || schemes.length === 0 || schemes.some((scheme) => typeof scheme !== "string" || scheme === "")) {
    throw new Error("redaction manifest has no credential URL schemes");
  }
  let text = one(values, "text") ?? "";
  for (const secret of secrets) text = text.replaceAll(secret, marker);
  const schemeAlternatives = schemes.map((scheme) => scheme.replace(/[\\^$.*+?()[\]{}|]/gu, "\\$&")).join("|");
  // Explicit ASCII classes, as in every shared redaction pattern (TH-HARNESS-REQ-617).
  text = text.replace(new RegExp(`(?:${schemeAlternatives}):\\/\\/[^\\t\\n\\f\\r ]+`, "giu"), marker);
  text = text.replace(/[\u0000-\u001f\u007f]+/gu, " ").trim();
  if (text.length > 1500) text = text.slice(-1500);
  process.stdout.write(`${label} process failed to establish a session: ${text}\n`);
}

function writeSession([envFile, leaseFile]) {
  const values = records();
  const env = {}, lease = {};
  for (const [key, all] of values) {
    const value = all.at(-1);
    if (key.startsWith("env.")) {
      if (value !== "") env[key.slice(4)] = value;
    } else if (key.startsWith("lease.")) {
      lease[key.slice(6)] = value;
    } else throw new Error(`unknown session record ${key}`);
  }
  const document = {
    schema_id: "cartulary.web_e2e_session_lease.v1",
    env,
    session_env_file: envFile,
    backend_port: Number.parseInt(lease.backend_port ?? "", 10),
    frontend_port: Number.parseInt(lease.frontend_port ?? "", 10),
    runtime_root: lease.runtime_root,
    server_log: lease.server_log,
    web_log: lease.web_log,
    startup_diagnostics: lease.startup_diagnostics,
    frontend_mode: lease.frontend_mode,
    frontend_command_kind: lease.frontend_command_kind,
    server_pgid: lease.server_pgid,
    vite_pgid: lease.vite_pgid,
    keep_runtime_root: lease.keep_runtime_root === "1",
    e2e_db: lease.e2e_db,
    test_services_metadata_file: lease.test_services_metadata_file,
    runtime_profile_id: lease.runtime_profile_id,
    runtime_profile_fingerprint: lease.runtime_profile_fingerprint,
    backend_generation_head: lease.backend_generation_head,
    backend_restart_secret_file: lease.backend_restart_secret_file,
    backend_generation: 1,
  };
  for (const [key, value] of Object.entries(document)) if (value === undefined) delete document[key];
  writePrivate(envFile, env);
  writePrivate(leaseFile, document);
}

function updateLease([file]) {
  const values = records();
  const lease = JSON.parse(fs.readFileSync(file, "utf8"));
  lease.server_pgid = one(values, "server_pgid");
  lease.backend_generation = Number.parseInt(one(values, "backend_generation") ?? "", 10);
  lease.backend_generation_head = one(values, "backend_generation_head");
  lease.env.CARTULARY_WEB_E2E_BACKEND_GENERATION_HEAD = lease.backend_generation_head;
  writePrivate(file, lease);
}

const commands = {
  "write-restart-secrets": [writeRestartSecrets, 1],
  "failure-message": [failureMessage, 2],
  "write-session": [writeSession, 2],
  "update-lease": [updateLease, 1],
};

const [command, ...args] = process.argv.slice(2);
const entry = commands[command];
if (!entry || args.length !== entry[1] || args.some((arg) => arg === "")) {
  process.stderr.write("usage: web-e2e-session-cli.mjs write-restart-secrets <file> | failure-message <label> <manifest> | write-session <env-file> <lease-file> | update-lease <lease-file>\n");
  process.exit(2);
}
entry[0](args);
