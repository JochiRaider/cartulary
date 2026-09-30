import { coreReadiness, nodeReadiness } from "../../readiness/installed-readiness.mjs";
import { preparationFailure } from "./failure.mjs";
import { repoRoot } from "./policy.mjs";

export function controlProfile(root = repoRoot) {
  try { return { node_version: nodeReadiness(root).node_version }; }
  catch (cause) { throw preparationFailure(cause); }
}
export function toolProfile(root = repoRoot) {
  try { return coreReadiness(root); }
  catch (cause) { throw preparationFailure(cause); }
}
if (process.argv[2] === "--doctor") {
  try { toolProfile(); process.stdout.write("ok UI review core toolchain\n"); }
  catch (error) { process.stderr.write(`missing UI review prerequisite: ${error.context.subject_id}; recovery=${error.context.recovery_id}\n`); process.exitCode = 2; }
}
