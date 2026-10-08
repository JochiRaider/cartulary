import { createHash, randomBytes } from "node:crypto";
import { lstatSync, mkdirSync, unlinkSync } from "node:fs";
import path from "node:path";
import os from "node:os";
import { generateRunId, validateResultRoot, validateRunId } from "../../contract/harness-contract.mjs";
import { parseStrictJSON } from "../../contract/index.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile } from "../../runtime/secure-local-files.mjs";
import { processIdentity, processIdentityAlive } from "../../runtime/host-admission.mjs";
import { restrictedExecutableInputRoots } from "../../test-catalog/index.mjs";
import { repoRoot } from "./policy.mjs";
import { ReviewFailure, validate, limits } from "./contract.mjs";

export const digest = (bytes) => createHash("sha256").update(bytes).digest("hex");
export const jsonBytes = (value) => Buffer.from(`${JSON.stringify(value)}\n`);
const registry = path.join(os.tmpdir(), `cartulary-ui-review-${process.getuid()}`);
export function inputPath(file, extension) {
  if (typeof file !== "string" || !file || /^[a-z][a-z0-9+.-]*:/iu.test(file) || /[\\\u0000-\u001f]/u.test(file) || file.split("/").includes("..")) throw new ReviewFailure("invalid_request");
  const absolute = path.resolve(repoRoot, file);
  const relative = path.relative(repoRoot, absolute).split(path.sep).join("/");
  const restricted = restrictedExecutableInputRoots(repoRoot);
  if (restricted.some((root) => relative === root || relative.startsWith(`${root}/`)) || /(?:^|\/)readme(?:\.[^/]*)?$|\.(?:md|markdown|mdown)$/iu.test(absolute) || (extension && path.extname(absolute).toLowerCase() !== extension)) throw new ReviewFailure("input_boundary");
  return absolute;
}
export function readInput(file, options) {
  try { return readLocalFile(inputPath(file, options.extension), options); }
  catch (cause) { if (cause instanceof ReviewFailure) throw cause; throw new ReviewFailure("unsafe_artifact", { cause }); }
}
export function readLocator(file) {
  const absolute = inputPath(file, ".json");
  try {
    const locator = validate("session", parseStrictJSON(readLocalFile(absolute, { maximum: limits.request }).toString("utf8")));
    if (path.basename(absolute) !== "session.json" || path.basename(path.dirname(absolute)) !== "ui-review" || path.basename(path.dirname(path.dirname(absolute))) !== locator.run_id) throw new Error("locator identity");
    return { locator, file: absolute, runRoot: path.dirname(path.dirname(absolute)) };
  } catch (cause) { throw new ReviewFailure("session_mismatch", { cause }); }
}
export function newRunRoot(environment = process.env) {
  try {
    const results = validateResultRoot(environment.CARTULARY_TEST_RESULTS_DIR || undefined, { root: repoRoot, create: false });
    const runID = environment.CARTULARY_TEST_RUN_ID ? validateRunId(environment.CARTULARY_TEST_RUN_ID) : generateRunId();
    privateDirectory(results);
    const runRoot = path.join(results, runID); mkdirSync(runRoot, { mode: 0o700 });
    privateDirectory(path.join(runRoot, "ui-review"));
    return { runRoot, runID };
  } catch (cause) { throw new ReviewFailure("session_mismatch", { cause }); }
}
export function publishJSON(file, name, value, { replace = false } = {}) {
  validate(name, value);
  const bytes = jsonBytes(value); atomicLocalFile(file, bytes, { replace });
  return { bytes: bytes.length, sha256: digest(bytes), media_type: "application/json" };
}
export function registerSession({ locatorFile, sessionID, runtime }) {
  privateDirectory(registry);
  const stem = sessionID.slice("uireview-".length);
  const record = { session_id: sessionID, locator: locatorFile, repository: repoRoot, process: processIdentity(), started_tick_ns: process.hrtime.bigint().toString(), token: randomBytes(32).toString("hex"), socket: path.join(registry, `${stem}.sock`), runtime: { root: runtime.root, lease_id: runtime.leaseID, run_id: runtime.runID } };
  atomicLocalFile(path.join(registry, `${stem}.json`), jsonBytes(record));
  return record;
}
export function resolveSession(identity, { allowDead = false, allowAbsent = false } = {}) {
  try {
    const file = path.join(registry, `${identity.locator.session_id.slice(9)}.json`);
    let bytes;
    try { bytes = readLocalFile(file); }
    catch (error) { if (allowAbsent && error.code === "ENOENT") return null; throw error; }
    const record = JSON.parse(bytes);
    if (!record.process || Object.keys(record.process).sort().join(",") !== "boot,pid,start" || !/^[a-f0-9-]{36}$/u.test(record.process.boot) || !Number.isSafeInteger(record.process.pid) || record.process.pid < 2 || !/^\d+$/u.test(record.process.start) || !/^\d+$/u.test(record.started_tick_ns) || !record.runtime || Object.keys(record.runtime).sort().join(",") !== "lease_id,root,run_id") throw new Error("invalid ownership identity");
    if (record.repository !== repoRoot || record.session_id !== identity.locator.session_id || record.locator !== identity.file || record.runtime.run_id !== identity.locator.run_id || !/^[0-9a-f]{64}$/u.test(record.token) || record.socket !== path.join(registry, `${identity.locator.session_id.slice(9)}.sock`)) throw new Error("lease mismatch");
    if (!processIdentityAlive(record.process)) {
      if (!allowDead) throw new Error("dead owner");
      return { ...record, dead: true };
    }
    const socket = lstatSync(record.socket);
    if (!socket.isSocket() || socket.uid !== process.getuid() || (socket.mode & 0o777) !== 0o600) throw new Error("socket ownership");
    return record;
  } catch (cause) { throw new ReviewFailure("session_mismatch", { cause }); }
}
export function unregisterSession(record) {
  for (const file of [record.socket, path.join(registry, `${record.session_id.slice(9)}.json`), `${record.socket}.recovery-lock`]) {
    try { unlinkSync(file); } catch (error) { if (error.code !== "ENOENT") throw error; }
  }
}
export function updateSessionRecord(record, fields) {
  Object.assign(record, fields);
  const { dead: _dead, ...persisted } = record;
  atomicLocalFile(path.join(registry, `${record.session_id.slice(9)}.json`), jsonBytes(persisted), { replace: true });
}
export function terminalResult(identity, command, makeResult) {
  const ref = identity.locator.terminal_receipt;
  if (!ref) throw new ReviewFailure("unsafe_artifact");
  let receipt;
  if (ref) {
    if (ref.path !== "ui-review/terminal.json" || ref.media_type !== "application/json") throw new ReviewFailure("invalid_artifact");
    const file = path.join(identity.runRoot, ref.path);
    if (path.relative(identity.runRoot, file).startsWith("..")) throw new ReviewFailure("invalid_artifact");
    let bytes;
    try { bytes = readLocalFile(file); }
    catch (cause) { throw new ReviewFailure("unsafe_artifact", { cause }); }
    if (bytes.length !== ref.bytes || digest(bytes) !== ref.sha256) throw new ReviewFailure("invalid_artifact");
    try { receipt = validate("receipt", JSON.parse(bytes)); }
    catch (cause) { throw new ReviewFailure("invalid_artifact", { cause }); }
    if (receipt.session_id !== identity.locator.session_id || receipt.state !== identity.locator.state) throw new ReviewFailure("invalid_artifact");
  }
  return makeResult(receipt, ref);
}
