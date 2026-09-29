import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync, unlinkSync } from "node:fs";
import path from "node:path";
import { atomicLocalFile, privateDirectory, readLocalFile, removePrivateFile, removePrivateTree } from "../../runtime/secure-local-files.mjs";
import { processIdentity, processIdentityAlive } from "../../runtime/host-admission.mjs";
import { ReviewFailure } from "./failure.mjs";

const kinds = ["browser_stack", "managed_suite", "preparation_process", "browser_process", "helper_process"];
const key = (kind, target) => `${kind}-${createHash("sha256").update(JSON.stringify(target)).digest("hex")}.json`;
export function recordResource(runtime, { kind, target, state = "acquired" }) {
  if (!kinds.includes(kind) || !["pending", "acquired", "released"].includes(state)) throw new ReviewFailure("unsafe_artifact");
  const directory = runtime.privatePath("recovery"); privateDirectory(directory);
  const file = path.join(directory, key(kind, target));
  if (state === "released") { try { unlinkSync(file); } catch (error) { if (error.code !== "ENOENT") throw error; } return; }
  const value = { kind, target, state, lease_id: runtime.leaseID, run_id: runtime.runID };
  atomicLocalFile(file, `${JSON.stringify(value)}\n`, { replace: existsSync(file) });
}
export function recoveryResources(runtime) {
  const directory = runtime.privatePath("recovery");
  let names;
  try { names = readdirSync(directory); } catch (error) { if (error.code === "ENOENT") return []; throw error; }
  if (names.length > 64) throw new ReviewFailure("unsafe_artifact");
  return names.map((name) => {
    const value = JSON.parse(readLocalFile(path.join(directory, name)));
    if (Object.keys(value).sort().join(",") !== "kind,lease_id,run_id,state,target" || !kinds.includes(value.kind) || !["pending", "acquired"].includes(value.state) || value.lease_id !== runtime.leaseID || value.run_id !== runtime.runID || name !== key(value.kind, value.target)) throw new ReviewFailure("unsafe_artifact");
    if (value.kind.endsWith("_process")) {
      const proof = value.target;
      if (!proof || Object.keys(proof).sort().join(",") !== "boot,group,pid,start" || !Number.isSafeInteger(proof.pid) || proof.pid < 2 || typeof proof.group !== "boolean" || typeof proof.boot !== "string" || !/^\d+$/u.test(proof.start)) throw new ReviewFailure("unsafe_artifact");
    } else {
      const relative = typeof value.target === "string" ? path.relative(runtime.root, value.target) : "..";
      if (!relative || relative.startsWith("..") || path.isAbsolute(relative) || relative.split(path.sep).some((part) => part === "recovery" || part === ".")) throw new ReviewFailure("unsafe_artifact");
    }
    return value;
  });
}
export function ownedProcess(pid) {
  const proof = processIdentity(pid);
  const fields = readFileSync(`/proc/${pid}/stat`, "utf8").split(") ").at(-1).split(" ");
  return { ...proof, group: Number(fields[2]) === pid };
}
export async function stopOwnedProcess(proof, { graceMS = 1000, killMS = 5000 } = {}) {
  // A detached process group can outlive its leader. Linux reserves its group
  // identifier until the last member exits; reject a reused leader identity.
  const alive = () => {
    if (!proof.group) return processIdentityAlive(proof);
    if (proof.boot !== processIdentity().boot) return false;
    try { if (processIdentity(proof.pid).start !== proof.start) return false; }
    catch (error) { if (!["ENOENT", "ESRCH"].includes(error.code)) throw error; }
    return readdirSync("/proc").some((pid) => {
      if (!/^\d+$/u.test(pid)) return false;
      try {
        const stat = readFileSync(`/proc/${pid}/stat`, "utf8");
        const fields = stat.slice(stat.lastIndexOf(")") + 2).split(" ");
        return !["Z", "X"].includes(fields[0]) && Number(fields[2]) === proof.pid && Number(fields[3]) === proof.pid && BigInt(fields[19]) >= BigInt(proof.start);
      } catch (error) { if (["ENOENT", "ESRCH"].includes(error.code)) return false; throw error; }
    });
  };
  if (!alive()) return;
  const signal = (name) => {
    if (!alive()) return;
    try { process.kill(proof.group ? -proof.pid : proof.pid, name); } catch (error) { if (error.code !== "ESRCH") throw error; }
  };
  const wait = async (duration) => {
    const end = performance.now() + duration;
    while (alive() && performance.now() < end) await new Promise((resolve) => setTimeout(resolve, 20));
    return !alive();
  };
  signal("SIGTERM"); if (await wait(graceMS)) return;
  signal("SIGKILL"); if (!await wait(killMS)) throw new ReviewFailure("cleanup_failed");
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
