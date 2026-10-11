import { spawn, execFileSync } from "../workspace/child-process.mjs";
import { randomBytes } from "node:crypto";
import { readFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
const nativeAdmission = path.resolve(import.meta.dirname, "../../..", workspaceLayout.toolbin, "cartulary-harness-admission");

const bootID = readFileSync("/proc/sys/kernel/random/boot_id", "utf8").trim();
export function processIdentity(pid = process.pid) {
  const stat = readFileSync(`/proc/${pid}/stat`, "utf8");
  return { boot: bootID, pid, start: stat.slice(stat.lastIndexOf(")") + 2).split(" ")[19] };
}
export function processIdentityAlive(proof) {
  if (!proof || proof.boot !== bootID || !Number.isSafeInteger(proof.pid) || !/^[0-9]+$/u.test(proof.start)) return false;
  try {
    const stat = readFileSync(`/proc/${proof.pid}/stat`, "utf8");
    const fields = stat.slice(stat.lastIndexOf(")") + 2).split(" ");
    return !["Z", "X"].includes(fields[0]) && fields[19] === proof.start;
  }
  catch (error) { if (error.code === "ENOENT" || error.code === "ESRCH") return false; throw error; }
}
function conflict() { return Object.assign(new Error("host admission unavailable"), { failure_class: "infra", failure_reason: "resource_conflict" }); }
export const hostAdmissionRoot = path.join(os.tmpdir(), `cartulary-host-admission-${process.getuid()}`);
const delay = (milliseconds, signal) => new Promise((resolve, reject) => {
  signal?.throwIfAborted();
  const abort = () => { clearTimeout(timer); reject(signal.reason); };
  const timer = setTimeout(() => { signal?.removeEventListener("abort", abort); resolve(); }, milliseconds);
  signal?.addEventListener("abort", abort, { once: true });
});
async function transaction(root, request, timeoutMs = 5000) {
  const child = spawn(nativeAdmission, [], { stdio: ["ignore", "pipe", "ignore"], env: {
    CARTULARY_HOST_ADMISSION_REQUEST: JSON.stringify({ root, ...request }),
    CARTULARY_HOST_ADMISSION_TIMEOUT_MS: String(Math.max(1, Math.floor(timeoutMs))),
  } });
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => child.kill("SIGKILL"), Math.max(1, timeoutMs));
    let output = "";
    child.once("error", (error) => { clearTimeout(timer); reject(error); });
    child.stdout.on("data", (bytes) => { output += bytes; if (output.length > 1024) child.kill("SIGKILL"); });
    child.once("close", (status) => {
      clearTimeout(timer);
      try { const result = JSON.parse(output); if (status !== 0 || result.ok !== true) throw conflict(); resolve(result.value); }
      catch { reject(conflict()); }
    });
  });
}
// Executed entirely inside the short-lived native transaction child. A synchronous build or
// fixture operation in the scheduler must never hold the mutex awaiting its
// parent's event loop. The child publishes and exits without a parent handshake.
const resources = ["browser_stack", "cpu", "io", "memory_mb", "process"];
const quantities = (value) => value && Object.keys(value).length === resources.length && resources.every((key) => Number.isSafeInteger(value[key]) && value[key] >= 0);
export const hostClaims = (value = {}) => Object.fromEntries(resources.map((key) => [key, value[key] ?? 0]));
export function inheritedHostLease(environment = process.env) {
  const token = environment.CARTULARY_HOST_ADMISSION_LEASE;
  if (token === undefined) return null;
  if (!/^[a-f0-9]{32}$/u.test(token)) throw conflict();
  return token;
}
// Synchronous transaction access is retained for controlled owner fixtures;
// runtime callers use the asynchronous adapter below. Both execute one owner.
export function transactAdmission(request) {
  try {
    const result = JSON.parse(execFileSync(nativeAdmission, [], { encoding: "utf8", timeout: 5000, maxBuffer: 1024,
      env: { CARTULARY_HOST_ADMISSION_REQUEST: JSON.stringify(request) } }));
    if (result.ok !== true) throw conflict();
    return result.value;
  } catch { throw conflict(); }
}
/** One host-wide queue and weighted accounting tree for scheduled and review work.
 * Children inherit activity order but consume only unused parent claim credit.
 * Exact worker proof retains claims after controller death until reaping.
 */
export async function acquireHostAdmission({ mode = "shared", claims: input = {}, capacities: available, parent = null, signal, timeoutMs = 5000, root = hostAdmissionRoot, ownerIdentity = processIdentity() } = {}) {
  const claims = hostClaims(input), capacities = hostClaims(available);
  if (!["shared", "exclusive"].includes(mode) || !quantities(claims) || !quantities(capacities) || resources.some((key) => claims[key] > capacities[key])) throw conflict();
  const token = randomBytes(16).toString("hex"), owner = ownerIdentity, started = performance.now();
  let registered = false;
  const release = async () => { if (registered) { await transaction(root, { operation: "release", token }); registered = false; } };
  try {
    while (true) {
      signal?.throwIfAborted();
      registered = true;
      const admitted = await transaction(root, { operation: "admit", token, owner, mode, claims, capacities, parent }, Math.max(1, timeoutMs - (performance.now() - started)));
      signal?.throwIfAborted();
      if (admitted) return { token, release, bind: (guard) => transaction(root, { operation: "bind", token, guard }) };
      if (performance.now() - started >= timeoutMs) throw conflict();
      await delay(Math.min(50, Math.max(1, timeoutMs - (performance.now() - started))), signal);
    }
  } catch (error) { try { await release(); } catch (secondary) { (error.cleanupFailures ??= []).push(secondary); } throw error; }
}
