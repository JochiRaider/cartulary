// Package-owned resource names; the native runtime owns journal and recovery order.
import { spawnSync } from "node:child_process";
import { randomUUID } from "node:crypto";
import { parseStrictJSON, validateSchemaSync } from "../harness/contract/index.mjs";
import { CommandFailure } from "../harness/runtime/command-failure.mjs";

function failed(cause) {
  return new CommandFailure("package resource cleanup failed", { failure_class: "harness", failure_reason: "cleanup_error" }, { cause });
}
function docker(args) {
  const result = spawnSync("docker", args, { timeout: 5000, maxBuffer: 1024 * 1024, killSignal: "SIGKILL" });
  if (result.error || result.status !== 0) throw failed(result.error);
  return new TextDecoder("utf-8", { fatal: true }).decode(result.stdout).trim();
}
export function packageResourceIdentity(resource) {
  validateSchemaSync("cartulary.harness_runtime_resource.v2", resource);
  if (resource.kind !== "package_resource") throw failed();
  const { token, role } = resource.target;
  return {
    name: role === "build_tag" ? `cartulary/package-build:${token}` : `cartulary-package-${token}-${role}`,
    labels: { "cartulary.package.allocation": token, "cartulary.package.run": resource.run_id,
      "cartulary.package.lease": resource.lease_id, "cartulary.package.role": role },
  };
}
function listing(resource) {
  const { name } = packageResourceIdentity(resource);
  return resource.target.role === "build_tag"
    ? ["image", "ls", "--no-trunc", "--quiet", "--filter", `reference=${name}`]
    : ["ps", "--all", "--no-trunc", "--filter", `name=^/${name}$`, "--format", "{{.ID}}"];
}

// Reservation precedes effects, including commands whose response can be lost.
// The journal, rather than a successful create response, supplies recovery authority.
export function reservePackageResources(runtime, run = docker) {
  const daemon_id = run(["info", "--format", "{{.ID}}"]), token = randomUUID();
  const resources = ["application", "postgres", "seaweedfs", "build_tag"].map(role => ({
    kind: "package_resource", target: { daemon_id, token, role }, state: "pending",
    run_id: runtime.runID, lease_id: runtime.leaseID, depends_on: [],
  }));
  for (const resource of resources) {
    if (run(listing(resource))) throw failed(); // Never reserve an existing name.
  }
  const active = new Map();
  function release(role) {
    const resource = active.get(role);
    if (!resource) return;
    cleanupPackageResource(resource, run);
    runtime.record({ ...resource, state: "released" });
    active.delete(role);
  }
  function releaseAll() {
    const failures = [];
    for (const role of active.keys()) {
      try { release(role); }
      catch (cause) { failures.push(Object.assign(failed(cause), { cleanup_step: role })); }
    }
    return failures;
  }
  try {
    for (const resource of resources) {
      // A failed journal reply can follow a durable write. Include this resource
      // in settlement even if record() does not return successfully.
      active.set(resource.target.role, resource);
      runtime.record(resource);
    }
  } catch (error) {
    const failures = releaseAll();
    if (failures.length) (error.cleanupFailures ??= []).push(...failures);
    throw error;
  }
  return Object.freeze({
    identity: role => {
      const resource = active.get(role);
      if (!resource) throw failed();
      return packageResourceIdentity(resource);
    }, release, releaseAll,
  });
}
export function cleanupPackageResource(resource, run = docker) {
  try {
    const identity = packageResourceIdentity(resource);
    if (run(["info", "--format", "{{.ID}}"]) !== resource.target.daemon_id) throw failed();
    const image = resource.target.role === "build_tag";
    const query = listing(resource);
    const id = run(query);
    if (!id) return;
    if (!(image ? /^sha256:[a-f0-9]{64}$/u : /^[a-f0-9]{64}$/u).test(id)) throw failed();
    if (!image) {
      const value = parseStrictJSON(run(["inspect", "--format", "{{json .}}", id]));
      if (value.Id !== id || value.Name !== `/${identity.name}` ||
          Object.entries(identity.labels).some(([key, expected]) => value.Config?.Labels?.[key] !== expected)) throw failed();
    }
    run(image ? ["image", "rm", identity.name] : ["rm", "--force", id]);
    if (run(query)) throw failed();
  } catch (cause) {
    if (cause instanceof CommandFailure) throw cause;
    throw failed(cause);
  }
}
