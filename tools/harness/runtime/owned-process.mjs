import { readFileSync, readdirSync } from "node:fs";
import { processIdentity, processIdentityAlive } from "./host-admission.mjs";

export function ownedProcess(pid) {
  const proof = processIdentity(pid);
  const fields = readFileSync(`/proc/${pid}/stat`, "utf8").split(") ").at(-1).split(" ");
  return { ...proof, group: Number(fields[2]) === pid };
}
export async function stopOwnedProcess(proof, { graceMS = 1000, killMS = 5000, exitNotification } = {}) {
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
  // Direct-child callers already receive an exit event. Use it as a wakeup,
  // never as proof that a process group is empty or that an identity is gone.
  let exitObserved = false;
  const exitWake = exitNotification && Promise.resolve(exitNotification).then(
    () => { exitObserved = true; }, () => { exitObserved = true; });
  const wait = async (duration) => {
    const end = performance.now() + duration;
    while (alive() && performance.now() < end) {
      let timer;
      const delay = new Promise((resolve) => { timer = setTimeout(resolve, Math.min(20, Math.max(0, end - performance.now()))); });
      try { await (exitWake && !exitObserved ? Promise.race([delay, exitWake]) : delay); }
      finally { clearTimeout(timer); }
    }
    return !alive();
  };
  signal("SIGTERM"); if (await wait(graceMS)) return;
  signal("SIGKILL"); if (!await wait(killMS)) throw new Error("owned process cleanup failed");
}

