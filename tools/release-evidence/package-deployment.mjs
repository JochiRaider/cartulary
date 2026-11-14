// Disposable deployment policy. Native admission and dependency ordering own
// authority; this adapter owns exact Docker observations and physical effects.
import { spawnSync } from "node:child_process";
import { closeSync, fstatSync } from "node:fs";
import { parseStrictJSON, validateSchemaSync } from "../harness/contract/index.mjs";
import { openDirectory, removePrivateTree } from "../harness/runtime/secure-local-files.mjs";
import { CommandFailure } from "../harness/runtime/command-failure.mjs";

export const packageDeploymentLimit = 128;
function failed(cause) {
 return new CommandFailure("package deployment cleanup failed", { failure_class: "harness", failure_reason: "cleanup_error" }, { cause });
}
export function packageDeploymentDocker(args, { timeout = 5000 } = {}) {
 const result = spawnSync("docker", args, { timeout, maxBuffer: 1024 * 1024, killSignal: "SIGKILL" });
 if (result.error || result.status !== 0) throw failed(result.error);
 return new TextDecoder("utf-8", { fatal: true }).decode(result.stdout).trim();
}
export function packageDeploymentIdentity(resource) {
 validateSchemaSync("cartulary.harness_runtime_resource.v2", resource);
 if (resource.kind !== "package_deployment") throw failed();
 const { token, scope } = resource.target;
 const prefix = `cartulary-qualification-${token}`;
 return { project: scope === "shared" ? null : `${prefix}-${scope}`, helper: `${prefix}-cleanup`,
  labels: { "cartulary.package.allocation": token, "cartulary.package.run": resource.run_id,
   "cartulary.package.lease": resource.lease_id, "cartulary.package.role": "qualification_cleanup" } };
}
function checkedList(value, pattern) {
 if (typeof value !== "string") throw failed();
 if (!value) return [];
 const values = value.split("\n");
 if (values.length > packageDeploymentLimit || new Set(values).size !== values.length || values.some(id => !pattern.test(id))) throw failed();
 return values;
}
const objectName = /^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,255}$/u;
const objectID = /^[a-f0-9]{64}$/u;
export function packageDeploymentQuery(resource) {
 const { project } = packageDeploymentIdentity(resource), { family } = resource.target;
 if (!project || family === "workspace") throw failed();
 return [family, "ls", ...(family === "container" ? ["--all", "--no-trunc"] : family === "network" ? ["--no-trunc"] : []),
  "--filter", `label=com.docker.compose.project=${project}`, "--format", family === "volume" ? "{{.Name}}" : "{{.ID}}"];
}
function inspect(run, args) {
 const value = parseStrictJSON(run(args));
 if (!Array.isArray(value) || value.length !== 1 || !value[0] || typeof value[0] !== "object") throw failed();
 return value[0];
}
function cleanupFamily(resource, run) {
 const { family } = resource.target, { project } = packageDeploymentIdentity(resource), query = packageDeploymentQuery(resource);
 const pattern = family === "volume" ? objectName : objectID;
 const ids = checkedList(run(query), pattern), failures = [];
 for (const id of ids) {
  try {
   const value = inspect(run, [family, "inspect", id]);
   const labels = family === "container" ? value.Config?.Labels : value.Labels;
   if ((family === "volume" ? value.Name : value.Id) !== id || labels?.["com.docker.compose.project"] !== project) throw failed();
   run([family, "rm", ...(family === "container" ? ["--force"] : []), id]);
  } catch (cause) { failures.push(cause); }
 }
 try { if (checkedList(run(query), pattern).length) throw failed(); } catch (cause) { failures.push(cause); }
 if (failures.length) throw failed(new AggregateError(failures));
}
function helperID(resource, run) {
 const { helper, labels } = packageDeploymentIdentity(resource);
 const ids = checkedList(run(["container", "ls", "--all", "--no-trunc", "--filter", `name=^/${helper}$`, "--format", "{{.ID}}"]), objectID);
 if (ids.length > 1) throw failed();
 if (!ids.length) return "";
 const value = inspect(run, ["container", "inspect", ids[0]]);
 if (value.Id !== ids[0] || value.Name !== `/${helper}` || value.Image !== resource.target.helper_image ||
     Object.entries(labels).some(([key, expected]) => value.Config?.Labels?.[key] !== expected)) throw failed();
 return ids[0];
}
function removeHelper(resource, run) {
 const id = helperID(resource, run);
 if (id) {
  run(["container", "rm", "--force", id]);
  if (helperID(resource, run)) throw failed();
 }
}
function cleanupWorkspace(resource, run) {
 const { workspace, helper_image } = resource.target, { helper, labels } = packageDeploymentIdentity(resource);
 // A previous attempt may have died during repair. Its helper is covered by
 // this pending workspace proof even when the create response was lost.
 removeHelper(resource, run);
 let descriptor;
 try { descriptor = openDirectory(workspace, { privateLeaf: true }); }
 catch (error) { if (error.code === "ENOENT") return; throw error; }
 let before, requested = false, primary;
 try {
  before = fstatSync(descriptor);
  // Docker's --mount uses CSV fields. Quote each complete field so private
  // paths containing commas or quotes cannot become additional mount options.
  const mount = ["type=bind", `source=${workspace}`, "target=/cleanup"].map(value => `"${value.replaceAll('"', '""')}"`).join(",");
  requested = true;
  const id = run(["create", "--name", helper, ...Object.entries(labels).flatMap(([key, value]) => ["--label", `${key}=${value}`]),
   "--network", "none", "--read-only", "--user", "0:0", "--cap-drop", "ALL", "--cap-add", "CHOWN", "--cap-add", "DAC_OVERRIDE",
   "--security-opt", "no-new-privileges:true", "--entrypoint", "sh", "--mount", mount, helper_image,
   "-c", 'chown -Rh "$1" /cleanup', "sh", `${process.getuid()}:${process.getgid()}`]);
  if (!objectID.test(id) || helperID(resource, run) !== id) throw failed();
  run(["start", "--attach", id], { timeout: 110000 });
 } catch (error) { primary = error; }
 finally {
  const errors = [];
  if (requested) { try { removeHelper(resource, run); } catch (error) { errors.push(error); } }
  try { closeSync(descriptor); } catch (error) { errors.push(error); }
  if (errors.length) {
   if (primary) (primary.cleanupFailures ??= []).push(...errors);
   else primary = failed(new AggregateError(errors));
  }
 }
 if (primary) throw primary;
 let current;
 try { current = openDirectory(workspace, { privateLeaf: true }); }
 catch (error) { if (error.code === "ENOENT") return; throw error; }
 try {
  const after = fstatSync(current);
  if (before.ino !== after.ino || before.dev !== after.dev) throw failed();
 } finally { closeSync(current); }
 removePrivateTree(workspace);
}
export function cleanupPackageDeployment(resource, run = packageDeploymentDocker) {
 try {
  packageDeploymentIdentity(resource);
  if (run(["info", "--format", "{{.ID}}"]) !== resource.target.daemon_id) throw failed();
  if (resource.target.family === "workspace") cleanupWorkspace(resource, run);
  else cleanupFamily(resource, run);
 } catch (cause) { if (cause instanceof CommandFailure) throw cause; throw failed(cause); }
}
