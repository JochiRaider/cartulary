import { createHash } from "node:crypto";
import { existsSync, readdirSync, unlinkSync } from "node:fs";
import { closeSync } from "node:fs";
import path from "node:path";
import { atomicLocalFile, privateDirectory, readLocalFile, openDirectory } from "./secure-local-files.mjs";

const kinds = ["browser_stack", "managed_suite", "object_store_proxy", "preparation_process", "browser_process", "helper_process", "diagnostic_scope"];
const key = (kind, target) => `${kind}-${createHash("sha256").update(JSON.stringify(target)).digest("hex")}.json`;
export function validateDiagnosticScope(scope) {
  if (!scope || Object.keys(scope).sort().join(",") !== "boot,start,token,uid" || typeof scope.boot !== "string" || !/^\d+$/u.test(scope.start) || !/^[a-f0-9]{64}$/u.test(scope.token) || scope.uid !== process.getuid()) throw new Error("unsafe runtime recovery proof");
}

export function recordRuntimeResource(runtime, { kind, target, state = "acquired" }) {
  if (!kinds.includes(kind) || !["pending", "acquired", "released"].includes(state)) throw new Error("unsafe runtime recovery proof");
  if (["browser_stack", "managed_suite", "object_store_proxy"].includes(kind)) {
    const relative = typeof target === "string" ? path.relative(runtime.root, target) : "..";
    if (!relative || relative.startsWith("..") || path.isAbsolute(relative) || relative.split(path.sep).includes("recovery")) throw new Error("unsafe runtime recovery target");
  }
  const directory = runtime.privatePath("recovery"); privateDirectory(directory);
  const file = path.join(directory, key(kind, target));
  if (state === "released") { try { unlinkSync(file); } catch (error) { if (error.code !== "ENOENT") throw error; } return; }
  const value = { kind, target, state, lease_id: runtime.leaseID, run_id: runtime.runID };
  atomicLocalFile(file, `${JSON.stringify(value)}\n`, { replace: existsSync(file) });
}
export function runtimeRecoveryResources(runtime, { maximum = 64 } = {}) {
  const directory = runtime.privatePath("recovery");
  let names;
  try { names = readdirSync(directory); } catch (error) { if (error.code === "ENOENT") return []; throw error; }
  if (names.length > maximum) throw new Error("unsafe runtime recovery proof");
  return names.map((name) => {
    const value = JSON.parse(readLocalFile(path.join(directory, name)));
    if (Object.keys(value).sort().join(",") !== "kind,lease_id,run_id,state,target" || !kinds.includes(value.kind) || !["pending", "acquired"].includes(value.state) || value.lease_id !== runtime.leaseID || value.run_id !== runtime.runID || name !== key(value.kind, value.target)) throw new Error("unsafe runtime recovery proof");
    if (value.kind === "diagnostic_scope") validateDiagnosticScope(value.target);
    else if (value.kind.endsWith("_process")) {
      const proof = value.target;
      if (!proof || Object.keys(proof).sort().join(",") !== "boot,group,pid,start" || !Number.isSafeInteger(proof.pid) || proof.pid < 2 || typeof proof.group !== "boolean" || typeof proof.boot !== "string" || !/^\d+$/u.test(proof.start)) throw new Error("unsafe runtime recovery proof");
    } else {
      const relative = typeof value.target === "string" ? path.relative(runtime.root, value.target) : "..";
      if (!relative || relative.startsWith("..") || path.isAbsolute(relative) || relative.split(path.sep).some((part) => part === "recovery" || part === ".")) throw new Error("unsafe runtime recovery proof");
      const parent = openDirectory(path.dirname(value.target), { privateLeaf: true }); closeSync(parent);
      if (value.state === "acquired") readLocalFile(value.target, { maximum: 1048576 });
    }
    return value;
  });
}
