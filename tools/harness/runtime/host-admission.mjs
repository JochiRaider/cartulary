import { spawn } from "node:child_process";
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
async function transaction(root, request) {
  privateDirectory(root);
  const lock = path.join(root, "mutex");
  try { atomicLocalFile(lock, ""); } catch (error) { if (error.code !== "EEXIST") throw error; }
  readLocalFile(lock, { maximum: 0 });
  const fd = openSync(lock, constants.O_RDONLY | constants.O_NOFOLLOW);
  const child = spawn("/usr/bin/flock", ["-x", "-w", "5", "/proc/self/fd/3", process.execPath, path.join(import.meta.dirname, "lease-lock.mjs")], { stdio: ["ignore", "pipe", "ignore", fd], env: { CARTULARY_HOST_ADMISSION_REQUEST: JSON.stringify({ root, ...request }) } });
  closeSync(fd);
  return new Promise((resolve, reject) => {
    let output = "";
    child.once("error", reject);
    child.stdout.on("data", (bytes) => { output += bytes; if (output.length > 1024) child.kill("SIGKILL"); });
    child.once("close", (status) => {
      try { const result = JSON.parse(output); if (status !== 0 || result.ok !== true) throw conflict(); resolve(result.value); }
      catch { reject(conflict()); }
    });
  });
}
// Executed entirely inside the short-lived flock child. A synchronous build or
// fixture operation in the scheduler must never hold the mutex awaiting its
// parent's event loop. The child publishes and exits without a parent handshake.
export function transactAdmission({ root, operation, token, owner, mode, browsers, browserCapacity }) {
  const file = path.join(root, "state.json");
  let state;
  try { state = JSON.parse(readLocalFile(file, { maximum: 4 * 1024 ** 2 })); } catch (error) { if (error.code !== "ENOENT") throw error; state = { sequence: 0, leases: [] }; }
  if (Object.keys(state).sort().join(",") !== "leases,sequence" || !Number.isSafeInteger(state.sequence) || state.sequence < 0 || !Array.isArray(state.leases) || state.leases.length > 4096) throw conflict();
  const tokens = new Set();
  for (const lease of state.leases) {
    if (Object.keys(lease).sort().join(",") !== "browsers,capacity,granted,mode,process,sequence,token" || !/^[a-f0-9]{32}$/u.test(lease.token) || tokens.has(lease.token) || !["shared", "exclusive"].includes(lease.mode) || typeof lease.granted !== "boolean" || !Number.isSafeInteger(lease.sequence) || lease.sequence < 1 || lease.sequence > state.sequence || !Number.isInteger(lease.browsers) || lease.browsers < 0 || !Number.isInteger(lease.capacity) || lease.capacity < lease.browsers) throw conflict();
    tokens.add(lease.token);
  }
  state.leases = state.leases.filter((lease) => processIdentityAlive(lease.process));
  let value = false;
  if (operation === "release") state.leases = state.leases.filter((lease) => lease.token !== token);
  else if (operation === "admit") {
    let lease = state.leases.find((entry) => entry.token === token);
    if (!lease) {
      if (state.leases.length >= 4096 || !processIdentityAlive(owner)) throw conflict();
      lease = { token, process: owner, sequence: ++state.sequence, mode, browsers, capacity: browserCapacity, granted: false }; state.leases.push(lease);
    }
    const active = state.leases.filter((entry) => entry.granted && entry.token !== token);
    const predecessors = state.leases.filter((entry) => entry.sequence < lease.sequence && !entry.granted);
    const capacity = Math.min(browserCapacity, ...active.filter((entry) => entry.browsers > 0).map((entry) => entry.capacity));
    const fits = browsers === 0 || active.reduce((sum, entry) => sum + entry.browsers, browsers) <= capacity;
    const compatible = mode === "exclusive" ? active.length === 0 && predecessors.length === 0 : !active.some((entry) => entry.mode === "exclusive") && !predecessors.some((entry) => entry.mode === "exclusive");
    if (fits && compatible) lease.granted = true;
    value = lease.granted;
  } else throw conflict();
  atomicLocalFile(file, `${JSON.stringify(state)}\n`, { replace: true });
  return value;
}
/** One host-wide admission domain, shared by scheduled work and local review.
 * Leases retain only opaque tokens, process proofs, queue order and capacities.
 * A quiet waiter bars newer shared requests. Kernel mutex release survives death.
 */
export async function acquireHostAdmission({ mode = "shared", browsers = 0, browserCapacity, signal, timeoutMs = 5000, root = hostAdmissionRoot } = {}) {
  if (!["shared", "exclusive"].includes(mode) || !Number.isInteger(browsers) || browsers < 0 || !Number.isInteger(browserCapacity) || browserCapacity < browsers) throw conflict();
  const token = randomBytes(16).toString("hex");
  const owner = processIdentity();
  const started = performance.now();
  let registered = false;
  const release = async () => { if (registered) { await transaction(root, { operation: "release", token }); registered = false; } };
  try {
    while (true) {
      signal?.throwIfAborted();
      registered = true;
      const admitted = await transaction(root, { operation: "admit", token, owner, mode, browsers, browserCapacity });
      if (admitted) return { token, release };
      if (performance.now() - started >= timeoutMs) throw conflict();
      await delay(50, signal);
    }
  } catch (error) { await release(); throw error; }
}
