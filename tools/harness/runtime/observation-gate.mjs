import { readlinkSync, watch } from "node:fs";
import path from "node:path";
import { readLocalFile } from "./secure-local-files.mjs";
import { acquireHostAdmission, processIdentity, processIdentityAlive } from "./host-admission.mjs";

// Admission storage and worker-thread liveness belong to the runtime owner.
export function createObservationGate({ capacities, root }) {
  const ownerIdentity = processIdentity(Number(readlinkSync("/proc/thread-self").split("/").at(-1)));
  let watcher, timer;
  const close = () => { watcher?.close(); watcher = null; clearTimeout(timer); };
  function quiet() {
    try {
      const state = JSON.parse(readLocalFile(path.join(root, "state.json"), { maximum: 4 * 1024 ** 2, allowReplacement: true }));
      return state.leases.some((lease) => lease.mode === "exclusive" && !lease.released && processIdentityAlive(lease.process));
    } catch (error) { return error.code !== "ENOENT"; }
  }
  return {
    quiet,
    acquire: ({ signal }) => acquireHostAdmission({ mode: "shared", claims: {}, capacities,
      timeoutMs: 250, signal, ownerIdentity, root }),
    waitUntilAvailable(ready, failed) {
      close();
      const check = () => { if (!quiet()) { close(); ready(); } };
      try {
        watcher = watch(root, { persistent: false }, (_event, name) => {
          if (name !== "state.json") return;
          clearTimeout(timer); timer = setTimeout(check, 50);
        });
        watcher.on("error", () => { close(); failed(); });
        timer = setTimeout(check, 50);
      } catch { close(); failed(); }
      return close;
    },
    close,
  };
}
