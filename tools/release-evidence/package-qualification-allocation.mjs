import { randomUUID } from "node:crypto";
import { lstatSync } from "node:fs";
import path from "node:path";
import { privateDirectory } from "../harness/runtime/secure-local-files.mjs";
import { runPrivateCapturedProcess } from "../harness/runtime/private-child-process.mjs";
import { lifecyclePolicy } from "../harness/contract/generated/lifecycle-policy.mjs";
import { CommandFailure } from "../harness/runtime/command-failure.mjs";
import { packageDeploymentDocker, packageDeploymentIdentity, packageDeploymentQuery } from "./package-deployment.mjs";

function failed(step, cause) {
 return Object.assign(new CommandFailure("package qualification allocation cleanup failed", {
  failure_class: "harness", failure_reason: "cleanup_error",
 }, { cause }), { cleanup_step: step });
}
export function reservePackageDeployment({ runtime, repoRoot, runRoot, helperImage, environment = process.env, run = packageDeploymentDocker }) {
 const daemon_id = run(["info", "--format", "{{.ID}}"]), token = randomUUID();
 const workspace = runtime.producerPath("package-qualification", token);
 try { lstatSync(workspace); throw failed("admission"); } catch (error) { if (error.code !== "ENOENT") throw error; }
 const resource = target => ({ kind: "package_deployment", target: { daemon_id, token, ...target },
  state: "pending", run_id: runtime.runID, lease_id: runtime.leaseID, depends_on: [] });
 const work = resource({ scope: "shared", family: "workspace", workspace, helper_image: helperImage });
 const families = ["source", "destination"].flatMap(scope => ["volume", "network", "container"].map(family => resource({ scope, family })));
 for (const candidate of [work, ...families]) packageDeploymentIdentity(candidate);
 for (const candidate of families) {
  const query = packageDeploymentQuery(candidate);
  if (run(query)) throw failed("admission");
  const byName = [...query]; byName[byName.indexOf("--filter") + 1] = `name=${packageDeploymentIdentity(candidate).project}`;
  if (run(byName)) throw failed("admission");
 }
 if (run(["container", "ls", "--all", "--no-trunc", "--filter", `name=${packageDeploymentIdentity(work).helper}`, "--format", "{{.ID}}"])) throw failed("admission");
 const active = new Map(), references = new Map();
 const key = value => `${value.target.scope}.${value.target.family}`;
 function record(candidate) {
  const reference = runtime.record(candidate);
  active.set(key(candidate), candidate); references.set(key(candidate), reference);
 }
 // No workspace or external acquisition occurs until every intent is durable.
 // An admission transport failure preserves uncertain proof for the native
 // owner; it cannot manufacture absence or erase a possibly durable write.
 record(work);
 for (const candidate of families.filter(value => value.target.family !== "container")) record(candidate);
 for (const candidate of families.filter(value => value.target.family === "container")) {
  candidate.depends_on = [references.get(`${candidate.target.scope}.volume`), references.get(`${candidate.target.scope}.network`), references.get("shared.workspace")];
  record(candidate);
 }
 privateDirectory(workspace);
 const env = { ...environment, ...runtime.producerEnvironment(), CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT: runRoot };
 async function retire(id) {
  const candidate = active.get(id);
  if (!candidate) return;
  let result, primary;
  try {
   result = await runPrivateCapturedProcess(process.execPath, [path.join(repoRoot, "tools/release-evidence/package-deployment-stop.mjs"), JSON.stringify(candidate)], {
    repoRoot, runRoot, cwd: repoRoot, env, signal: AbortSignal.timeout(lifecyclePolicy.cleanup_deadlines_ms.package_deployment),
    dependencies: [references.get(id)], captureLimitBytes: 1024 * 1024, tailBytes: 1024,
   });
   if (result.status !== 0 || result.signal) throw failed(id);
  } catch (error) { primary = error; }
  finally {
   if (result) { try { result.cleanup(); } catch (error) { if (primary) (primary.cleanupFailures ??= []).push(error); else primary = error; } }
  }
  if (primary) throw primary;
  runtime.record({ ...candidate, state: "released" }); active.delete(id);
 }
 return Object.freeze({ workspace,
  sourceProject: packageDeploymentIdentity(families.find(value => value.target.scope === "source")).project,
  destinationProject: packageDeploymentIdentity(families.find(value => value.target.scope === "destination")).project,
  dependencies: Object.freeze([references.get("source.container"), references.get("destination.container")]),
  releaseAll: async () => {
   const failures = [];
   const attempt = async id => { try { await retire(id); } catch (cause) { failures.push(failed(id, cause)); } };
   for (const scope of ["source", "destination"]) await attempt(`${scope}.container`);
   for (const scope of ["source", "destination"]) for (const family of ["volume", "network"]) {
    const id = `${scope}.${family}`;
    if (active.has(`${scope}.container`)) { if (active.has(id)) failures.push(failed(id, new Error("dependent container unresolved"))); }
    else await attempt(id);
   }
   if (active.has("source.container") || active.has("destination.container")) {
    if (active.has("shared.workspace")) failures.push(failed("shared.workspace", new Error("dependent container unresolved")));
   } else await attempt("shared.workspace");
   return failures;
  },
 });
}
