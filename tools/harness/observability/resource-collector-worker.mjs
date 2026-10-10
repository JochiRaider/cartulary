import { parentPort, workerData } from "node:worker_threads";
import { createObservationGate } from "../runtime/observation-gate.mjs";
import { createLinuxResourceAdapter, availabilityFor } from "./resource-linux.mjs";
import { createResourceStore } from "./resource-store.mjs";
import { createCollectionEngine } from "./collection-engine.mjs";

const { policy, identity, runRoot, epoch, capacities, admissionRoot } = workerData;
const cpuStart = process.threadCpuUsage();
let adapter;
try { adapter = createLinuxResourceAdapter(); } catch { /* Unsupported is retained separately from lifecycle. */ }
const engine = createCollectionEngine({ policy, identity, adapter, availabilityFor,
  gate: createObservationGate({ capacities, root: admissionRoot }),
  store: createResourceStore({ policy, runRoot }),
  clock: { now: () => performance.now(), elapsed: () => Number((process.hrtime.bigint() - BigInt(epoch)) / 1_000_000n), setTimeout, clearTimeout },
  observer: { cpu: () => process.threadCpuUsage(cpuStart), heap: () => process.memoryUsage().heapUsed },
  onRelationship: (record) => parentPort.postMessage({ type: "relationship", record }),
  onStatus: (value) => parentPort.postMessage({ type: "latest", value }),
});
parentPort.on("message", async (message) => {
  try {
    if (message.type === "register") { engine.register(message.proof, message.unit_id, message.allocation_ref, message.invocation_id); parentPort.postMessage({ type: "registered", request: message.request }); }
    if (message.type === "pause") { await engine.pause(); parentPort.postMessage({ type: "paused", request: message.request }); }
    if (message.type === "resume") await engine.resume();
    if (message.type === "stop") { await engine.stop(message.omitted_registrations); parentPort.postMessage({ type: "stopped" }); parentPort.close(); }
  } catch { parentPort.postMessage({ type: "failed" }); parentPort.close(); }
});
