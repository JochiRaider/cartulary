import { readdirSync } from "node:fs";
import path from "node:path";
import { removePrivateFile, removePrivateTree } from "../../runtime/secure-local-files.mjs";
import { stopOwnedProcess as stopProcess } from "../../runtime/owned-process.mjs";
import { ReviewFailure } from "./failure.mjs";
import { recordRuntimeResource, runtimeRecoveryResources } from "../../runtime/resource-recovery.mjs";

export { validateDiagnosticScope } from "../../runtime/resource-recovery.mjs";
export function recordResource(runtime, resource) {
  try { recordRuntimeResource(runtime, resource); }
  catch (cause) { throw new ReviewFailure("unsafe_artifact", { cause }); }
}
export function recoveryResources(runtime) {
  try { return runtimeRecoveryResources(runtime); }
  catch (cause) { throw new ReviewFailure("unsafe_artifact", { cause }); }
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
