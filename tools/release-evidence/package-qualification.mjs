// Package scenarios own observations; this controller owns their disposable
// allocation and joins the native workload before releasing its dependencies.
import { mkdirSync } from "node:fs";
import path from "node:path";
import { CommandFailure, createCommandFailureContext, reportCommandFailure } from "../harness/runtime/command-failure.mjs";
import { publicExitCodeForFailure } from "../harness/contract/index.mjs";
import { createLaunchContext } from "../harness/runtime/launch-context.mjs";
import { borrowRuntime } from "../harness/runtime/session-client.mjs";
import { runPrivateCapturedProcess } from "../harness/runtime/private-child-process.mjs";
import { atomicLocalFile } from "../harness/runtime/secure-local-files.mjs";
import { packageRunContext, readPackageReceipt } from "./package-identity.mjs";
import { reservePackageDeployment } from "./package-qualification-allocation.mjs";

const repoRoot = path.resolve(import.meta.dirname, "../..");
const qualifications = Object.freeze({
 package: { target: "standup-package-smoke", script: "check-standup-package-smoke.sh" },
 "reference-pack": { target: "standup-reference-pack-smoke", script: "check-standup-package-smoke.sh" },
 "credential-capacity": { target: "credential-capacity-assessment", script: "check-standup-package-smoke.sh" },
 recovery: { target: "standup-operational-recovery-smoke", script: "check-standup-operational-recovery-smoke.sh" },
});
function cleanupFailure(cause) {
 return new CommandFailure("package qualification cleanup failed", { failure_class: "harness", failure_reason: "cleanup_error" }, { cause });
}
export async function qualifyPackage(mode, { environment = process.env, workflowRoot = repoRoot } = {}) {
 const selected = Object.hasOwn(qualifications, mode) ? qualifications[mode] : null;
 if (!selected) throw new CommandFailure("invalid package qualification", { failure_class: "config", failure_reason: "usage_error" });
 const { runRoot } = packageRunContext(environment), receipt = readPackageReceipt(runRoot);
 const runtime = borrowRuntime({ repoRoot, runRoot, environment });
 const cancellation = new AbortController();
 const interrupt = () => cancellation.abort();
 process.on("SIGINT", interrupt); process.on("SIGTERM", interrupt);
 let allocation, result, primary, launch, diagnostic, settled = true;
 const artifactDirectory = path.join(runRoot, selected.target, "artifacts");
 try {
  allocation = reservePackageDeployment({ runtime, repoRoot, runRoot, environment, helperImage: receipt.images.find(value => value.role === "postgres").id });
  const installation = path.join(allocation.workspace, "installation"), temporary = path.join(allocation.workspace, "tmp");
  mkdirSync(installation, { mode: 0o700 }); mkdirSync(temporary, { mode: 0o700 });
  mkdirSync(artifactDirectory, { recursive: true, mode: 0o700 });
  const env = { ...environment, ...runtime.producerEnvironment(), CARTULARY_PACKAGE_TEMP_DIR: temporary,
   CARTULARY_PACKAGE_QUALIFICATION: mode, CARTULARY_PACKAGE_WORK_DIR: installation,
   CARTULARY_PACKAGE_SOURCE_PROJECT: allocation.sourceProject, CARTULARY_PACKAGE_DESTINATION_PROJECT: allocation.destinationProject,
   CARTULARY_PACKAGE_ARTIFACT_DIR: artifactDirectory };
  launch = createLaunchContext({ repoRoot, environment: env });
  diagnostic = launch.identity.command_id ? createCommandFailureContext({ repoRoot, environment: env, launch }) : null;
  settled = false;
  result = await runPrivateCapturedProcess("bash", [path.join(workflowRoot, "tools/release-evidence", selected.script)], {
    repoRoot, runRoot, cwd: repoRoot, env: { ...env, ...diagnostic?.environment }, launch, dependencies: allocation.dependencies, signal: cancellation.signal,
    captureLimitBytes: 16 * 1024 * 1024, tailBytes: 4096,
   });
  settled = true;
  if (cancellation.signal.aborted || result.signal) throw new CommandFailure("package qualification interrupted", { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" });
  const childFailure = diagnostic?.read();
  if (childFailure) throw new CommandFailure("classified package qualification failure", result.status === 0 ? { failure_class: "harness", failure_reason: "scheduler_accounting_error" } : childFailure);
  if (result.status !== 0) throw new CommandFailure("package qualification scenario failed", {
   failure_class: result.status === 2 ? "config" : "product", failure_reason: result.status === 2 ? "configuration_error" : "test_assertion_failure",
  });
 } catch (error) { primary = error; }
 finally {
  const failures = [];
  if (result) { try { result.cleanup(); } catch (error) { failures.push(cleanupFailure(error)); } }
  if (settled && diagnostic) { try { diagnostic.close(); } catch (error) { failures.push(cleanupFailure(error)); } }
  // Any failed capture may retain an unacknowledged journal write, even when
  // its physical child settled. Preserve its dependency edges for native recovery.
  // Native recovery must retire it before any deployment or workspace disposal.
  if (allocation) {
   if (settled) failures.push(...await allocation.releaseAll());
   else failures.push(cleanupFailure(new Error("workload settlement unresolved")));
   try { atomicLocalFile(path.join(artifactDirectory, "cleanup.json"), JSON.stringify({ cleanup: failures.length ? "failed" : "passed", failed_steps: failures.map(value => value.cleanup_step ?? "workload") }) + "\n"); }
   catch (error) { failures.push(cleanupFailure(error)); }
  }
  try { runtime.close(); } catch (error) { failures.push(cleanupFailure(error)); }
  process.off("SIGINT", interrupt); process.off("SIGTERM", interrupt);
  launch?.settle(cancellation.signal.aborted ? "cancelled" : primary || failures.length ? "failed" : "passed");
  if (failures.length) {
   if (primary) (primary.cleanupFailures ??= []).push(...failures);
   else primary = Object.assign(cleanupFailure(new AggregateError(failures)), { cleanupFailures: failures });
  }
 }
 if (primary) throw primary;
}
if (process.argv[1] && path.resolve(process.argv[1]) === import.meta.filename) {
 try {
  if (process.argv.length !== 3) throw new CommandFailure("invalid package qualification", { failure_class: "config", failure_reason: "usage_error" });
  await qualifyPackage(process.argv[2]);
 } catch (error) {
  const failure = reportCommandFailure(repoRoot, error, { failure_class: "artifact", failure_reason: "artifact_error" });
  console.error(`Package qualification failed: ${failure.failure_class}/${failure.failure_reason}`);
  if (error.cleanupFailures?.length) console.error(`Package qualification cleanup failed: ${error.cleanupFailures.length} steps`);
  process.exitCode = publicExitCodeForFailure(failure);
 }
}
