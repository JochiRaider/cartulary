import { createHash } from "node:crypto";
import { existsSync, readdirSync, unlinkSync } from "node:fs";
import path from "node:path";
import { atomicLocalFile, privateDirectory, readLocalFile, removePrivateFile, removePrivateTree } from "../../runtime/secure-local-files.mjs";
import { stopOwnedProcess as stopProcess } from "../../runtime/owned-process.mjs";
import { ReviewFailure } from "./failure.mjs";

const kinds = ["browser_stack", "managed_suite", "preparation_process", "browser_process", "helper_process", "diagnostic_scope"];
const key = (kind, target) => `${kind}-${createHash("sha256").update(JSON.stringify(target)).digest("hex")}.json`;
export function validateDiagnosticScope(scope) {
  if (!scope || Object.keys(scope).sort().join(",") !== "boot,start,token,uid" || typeof scope.boot !== "string" || !/^\d+$/u.test(scope.start) || !/^[a-f0-9]{64}$/u.test(scope.token) || scope.uid !== process.getuid()) throw new ReviewFailure("unsafe_artifact");
}

export function recordResource(runtime, { kind, target, state = "acquired" }) {
  if (!kinds.includes(kind) || !["pending", "acquired", "released"].includes(state)) throw new ReviewFailure("unsafe_artifact");
  const directory = runtime.privatePath("recovery"); privateDirectory(directory);
  const file = path.join(directory, key(kind, target));
  if (state === "released") { try { unlinkSync(file); } catch (error) { if (error.code !== "ENOENT") throw error; } return; }
  const value = { kind, target, state, lease_id: runtime.leaseID, run_id: runtime.runID };
  atomicLocalFile(file, `${JSON.stringify(value)}\n`, { replace: existsSync(file) });
}
export function recoveryResources(runtime) {
  const directory = runtime.privatePath("recovery");
  let names;
  try { names = readdirSync(directory); } catch (error) { if (error.code === "ENOENT") return []; throw error; }
  if (names.length > 64) throw new ReviewFailure("unsafe_artifact");
  return names.map((name) => {
    const value = JSON.parse(readLocalFile(path.join(directory, name)));
    if (Object.keys(value).sort().join(",") !== "kind,lease_id,run_id,state,target" || !kinds.includes(value.kind) || !["pending", "acquired"].includes(value.state) || value.lease_id !== runtime.leaseID || value.run_id !== runtime.runID || name !== key(value.kind, value.target)) throw new ReviewFailure("unsafe_artifact");
    if (value.kind === "diagnostic_scope") validateDiagnosticScope(value.target);
    else if (value.kind.endsWith("_process")) {
      const proof = value.target;
      if (!proof || Object.keys(proof).sort().join(",") !== "boot,group,pid,start" || !Number.isSafeInteger(proof.pid) || proof.pid < 2 || typeof proof.group !== "boolean" || typeof proof.boot !== "string" || !/^\d+$/u.test(proof.start)) throw new ReviewFailure("unsafe_artifact");
    } else {
      const relative = typeof value.target === "string" ? path.relative(runtime.root, value.target) : "..";
      if (!relative || relative.startsWith("..") || path.isAbsolute(relative) || relative.split(path.sep).some((part) => part === "recovery" || part === ".")) throw new ReviewFailure("unsafe_artifact");
    }
    return value;
  });
}
export { ownedProcess } from "../../runtime/owned-process.mjs";
export async function stopOwnedProcess(proof, options) {
  try { await stopProcess(proof, options); }
  catch (cause) { throw new ReviewFailure("cleanup_failed", { cause }); }
}

export async function boundedCleanup(release, milliseconds = 30000) {
  let timer;
  try {
    return await Promise.race([Promise.resolve().then(release), new Promise((_, reject) => { timer = setTimeout(() => reject(new ReviewFailure("cleanup_failed")), milliseconds); })]);
  } finally { clearTimeout(timer); }
}
export function purgeReviewDetail(runtime) {
  const failures = [];
  for (const file of ["suite-environment.json", "service-reaper.log"]) {
    try { removePrivateFile(runtime.privatePath("test-services", file)); } catch (error) { failures.push(error); }
  }
  for (const name of readdirSync(runtime.root)) {
    if (!["artifacts", "review-access", "lifecycle", "frontend-production"].includes(name) && !/^\.deleting-[a-f0-9]{32}$/u.test(name)) continue;
    try { removePrivateTree(runtime.privatePath(name)); } catch (error) { failures.push(error); }
  }
  if (failures.length) throw new AggregateError(failures, "private detail cleanup failed");
}
