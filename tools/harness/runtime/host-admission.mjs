import { spawn } from "../workspace/child-process.mjs";
import { randomBytes } from "node:crypto";
import { closeSync, constants, openSync, readFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { atomicLocalFile, privateDirectory, readLocalFile } from "./secure-local-files.mjs";

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
  privateDirectory(root);
  const lock = path.join(root, "mutex");
  try { atomicLocalFile(lock, ""); } catch (error) { if (error.code !== "EEXIST") throw error; }
  readLocalFile(lock, { maximum: 0 });
  const fd = openSync(lock, constants.O_RDONLY | constants.O_NOFOLLOW);
  const child = spawn("/usr/bin/flock", ["--no-fork", "-x", "-w", String(Math.max(0.001, timeoutMs / 1000)), "/proc/self/fd/3", process.execPath, path.join(import.meta.dirname, "lease-lock.mjs")], { stdio: ["ignore", "pipe", "ignore", fd], env: { CARTULARY_HOST_ADMISSION_REQUEST: JSON.stringify({ root, ...request }) } });
  closeSync(fd);
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
// Executed entirely inside the short-lived flock child. A synchronous build or
// fixture operation in the scheduler must never hold the mutex awaiting its
// parent's event loop. The child publishes and exits without a parent handshake.
const resources = ["browser_stack", "cpu", "io", "memory_mb", "process"];
const quantities = (value) => value && Object.keys(value).length === resources.length && resources.every((key) => Number.isSafeInteger(value[key]) && value[key] >= 0);
const proofShape = (value) => value && Object.keys(value).sort().join(",") === "boot,pid,start" && typeof value.boot === "string" && Number.isSafeInteger(value.pid) && value.pid > 1 && typeof value.start === "string" && /^[0-9]+$/u.test(value.start);
export const hostClaims = (value = {}) => Object.fromEntries(resources.map((key) => [key, value[key] ?? 0]));
export function inheritedHostLease(environment = process.env) {
  const token = environment.CARTULARY_HOST_ADMISSION_LEASE;
  if (token === undefined) return null;
  if (!/^[a-f0-9]{32}$/u.test(token)) throw conflict();
  return token;
}
export function transactAdmission({ root, operation, token, owner, mode, claims, capacities, parent, guard }) {
  const file = path.join(root, "state.json");
  let state;
  try { state = JSON.parse(readLocalFile(file, { maximum: 4 * 1024 ** 2 })); } catch (error) { if (error.code !== "ENOENT") throw error; state = { sequence: 0, leases: [] }; }
  if (Object.keys(state).sort().join(",") !== "leases,sequence" || !Number.isSafeInteger(state.sequence) || state.sequence < 0 || !Array.isArray(state.leases) || state.leases.length > 4096) throw conflict();
  const tokens = new Set();
  for (const lease of state.leases) {
    if (!proofShape(lease.process) || (!Array.isArray(lease.guards) || lease.guards.length > 64 || lease.guards.some((guard) => !proofShape(guard)))) throw conflict();
    // Nonempty obsolete ownership is deliberately not migrated or discarded.
    if (Object.keys(lease).sort().join(",") !== "capacities,claims,granted,guards,mode,parent,process,released,sequence,token" || !/^[a-f0-9]{32}$/u.test(lease.token) || tokens.has(lease.token) || !["shared", "exclusive"].includes(lease.mode) || typeof lease.granted !== "boolean" || typeof lease.released !== "boolean" || !Number.isSafeInteger(lease.sequence) || lease.sequence < 1 || lease.sequence > state.sequence || !quantities(lease.claims) || !quantities(lease.capacities) || (lease.parent !== null && !tokens.has(lease.parent))) throw conflict();
    tokens.add(lease.token);
  }
  const retain = new Set(state.leases.filter((lease) => (!lease.released && processIdentityAlive(lease.process)) || lease.guards.some(processIdentityAlive)).map((lease) => lease.token));
  for (const lease of [...state.leases].reverse()) if (retain.has(lease.token) && lease.parent) retain.add(lease.parent);
  state.leases = state.leases.filter((lease) => retain.has(lease.token));
  let value = false;
  if (operation === "release") {
    const lease = state.leases.find((entry) => entry.token === token);
    if (lease) lease.released = true;
    // A parent remains as an accounting node until every live child is gone.
    for (const entry of [...state.leases].reverse()) if (entry.released && !entry.guards.some(processIdentityAlive) && !state.leases.some((child) => child.parent === entry.token)) state.leases = state.leases.filter((item) => item !== entry);
  } else if (operation === "bind") {
    const lease = state.leases.find((entry) => entry.token === token);
    if (!lease?.granted || lease.released || !processIdentityAlive(guard)) throw conflict();
    lease.guards = lease.guards.filter(processIdentityAlive);
    if (!lease.guards.some((entry) => entry.boot === guard.boot && entry.pid === guard.pid && entry.start === guard.start)) lease.guards.push(guard);
    if (lease.guards.length > 64) throw conflict();
    value = true;
  } else if (operation === "admit") {
    if (!quantities(claims) || !quantities(capacities) || resources.some((key) => claims[key] > capacities[key])) throw conflict();
    let lease = state.leases.find((entry) => entry.token === token);
    if (!lease) {
      const ancestor = state.leases.find((entry) => entry.token === parent);
      if (state.leases.length >= 4096 || !processIdentityAlive(owner) || (parent && (!ancestor?.granted || ancestor.released))) throw conflict();
      lease = { token, process: owner, guards: [], sequence: ++state.sequence, mode, claims, capacities, parent: parent ?? null, granted: false, released: false }; state.leases.push(lease);
    }
    const rootOf = (entry) => entry.parent ? rootOf(state.leases.find((item) => item.token === entry.parent)) : entry;
    const rootLease = rootOf(lease), active = state.leases.filter((entry) => entry.granted);
    const predecessors = state.leases.filter((entry) => !entry.granted && entry.sequence < lease.sequence);
    const roots = active.filter((entry) => !entry.parent);
    const ancestors = new Set();
    for (let entry = lease; entry.parent;) { ancestors.add(entry.parent); entry = state.leases.find((item) => item.token === entry.parent); }
    const compatible = lease.parent
      ? !predecessors.some((entry) => entry.parent && rootOf(entry) === rootLease) && (mode === "exclusive" ? active.every((entry) => ancestors.has(entry.token)) : !active.some((entry) => entry.mode === "exclusive" && !ancestors.has(entry.token)))
      : mode === "exclusive" ? roots.length === 0 && predecessors.length === 0
        : !active.some((entry) => entry.mode === "exclusive") && !predecessors.some((entry) => !entry.parent || rootOf(entry) !== rootLease);
    // A subtree consumes the larger of its inherited claim and the sum of its
    // active children. Siblings cannot each borrow the same capacity credit.
    const included = new Set([...active, lease].map((entry) => entry.token));
    const total = (entry, key) => Math.max(entry.claims[key], state.leases.filter((child) => child.parent === entry.token && included.has(child.token)).reduce((sum, child) => sum + total(child, key), 0));
    const fits = resources.every((key) => {
      const capacity = Math.min(capacities[key], ...active.map((entry) => entry.capacities[key]));
      return state.leases.filter((entry) => !entry.parent && included.has(entry.token)).reduce((sum, entry) => sum + total(entry, key), 0) <= capacity;
    });
    if (fits && compatible) lease.granted = true;
    value = lease.granted;
  } else throw conflict();
  atomicLocalFile(file, `${JSON.stringify(state)}\n`, { replace: true });
  return value;
}
/** One host-wide queue and weighted accounting tree for scheduled and review work.
 * Children inherit activity order but consume only unused parent claim credit.
 * Exact worker proof retains claims after controller death until reaping.
 */
export async function acquireHostAdmission({ mode = "shared", claims: input = {}, capacities: available, parent = null, signal, timeoutMs = 5000, root = hostAdmissionRoot } = {}) {
  const claims = hostClaims(input), capacities = hostClaims(available);
  if (!["shared", "exclusive"].includes(mode) || !quantities(claims) || !quantities(capacities) || resources.some((key) => claims[key] > capacities[key])) throw conflict();
  const token = randomBytes(16).toString("hex"), owner = processIdentity(), started = performance.now();
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
