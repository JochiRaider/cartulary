import assert from "node:assert/strict";
import test from "node:test";
import { EventEmitter } from "node:events";
import { createDiagnosticSession } from "../diagnostic-session.mjs";
import { createResourceCollector } from "../resource-collector.mjs";
const deferred = () => { let resolve, reject; const promise = new Promise((yes, no) => { resolve = yes; reject = no; }); return { promise, resolve, reject }; };
const settle = async () => { for (let i = 0; i < 30; i += 1) await Promise.resolve(); };
const digest = `sha256:${"a".repeat(64)}`;
const identity = { run_id: "engine-fixture", source_digest: digest, graph_digest: digest, policy_digest: digest };
const proof = { identity: digest, pid: 1, start: "1", stat: { parent: 0 } };
for (const trigger of ["timeout", "error", "post_failure", "normal"]) test(`controller ${trigger} awaits worker termination; late acknowledgements cannot revive it`, async () => {
  const worker = new EventEmitter(), termination = deferred(), timers = new Map();
  let last, next = 0, terminated = false;
  worker.postMessage = (message) => { last = message; if (trigger === "post_failure" && message.type === "stop") throw new Error("transport"); };
  worker.terminate = () => { terminated = true; return termination.promise; };
  const collector = createResourceCollector({ runRoot: "unused", capacities: {}, epoch: process.hrtime.bigint(), manifest: { ...identity, instrumentation: { mode: "basic", policy_digest: digest } },
    workerFactory: () => worker, adapterFactory: () => ({ proof: () => proof }),
    timers: { setTimeout: (fn) => { timers.set(++next, fn); return next; }, clearTimeout: (id) => timers.delete(id) } });
  let done = false; const stop = collector.stop(); stop.then(() => { done = true; }); assert.equal(collector.stop(), stop);
  if (trigger === "timeout") timers.values().next().value();
  if (trigger === "error") worker.emit("error", new Error("worker death"));
  if (trigger === "normal") worker.emit("message", { type: "stopped" });
  await settle(); assert.equal(terminated, true); assert.equal(done, false);
  worker.emit("message", { type: "latest", value: { availability: "available" } });
  collector.resume(); assert.equal(last.type, "stop");
  termination.resolve(); await stop;
  assert.notEqual(collector.snapshot().availability, "available");
});

for (const reject of [false, true]) test(`diagnostic session closes live publication before scan eligibility, failure=${reject}`, async () => {
  const pending = deferred(), calls = [];
  const session = createDiagnosticSession({ mode: "off", epoch: 0n, elapsed: () => 1,
    collectorFactory: () => ({ stop: () => pending.promise }) });
  session.start({});
  const close = session.close(() => calls.push("live_closed"));
  assert.equal(session.close(), close); await settle(); assert.deepEqual(calls, []);
  if (reject) { pending.reject(new Error("cannot prove quiescence")); await assert.rejects(close, /quiescence/u); }
  else { pending.resolve(); await close; calls.push("scan_eligible"); }
  assert.deepEqual(calls, reject ? ["live_closed"] : ["live_closed", "scan_eligible"]);
});


import { spawn } from "node:child_process";
import { once } from "node:events";
import { ownedProcess, stopOwnedProcess } from "../../runtime/owned-process.mjs";

test("exit notifications wake cleanup without replacing live identity proof", async (t) => {
  const child = spawn(process.execPath, ["-e", 'process.on("SIGTERM",()=>process.stdout.write("stopping\\n"));process.stdin.on("data",()=>process.exit(0));process.stdout.write("ready\\n");setInterval(()=>{},1000)'], { stdio: ["pipe", "pipe", "ignore"] });
  t.after(() => { if (child.exitCode === null) child.kill("SIGKILL"); });
  const closed = once(child, "close");
  await once(child.stdout, "data");
  const proof = ownedProcess(child.pid);
  let settled = false;
  const stopping = once(child.stdout, "data");
  const completion = stopOwnedProcess(proof, { graceMS: 1000, exitNotification: Promise.resolve() });
  completion.then(() => { settled = true; });
  await stopping;
  await settle();
  assert.equal(settled, false, "an early event must not establish quiescence");
  child.stdin.end("finish");
  await completion; await closed;
  assert.equal(child.exitCode, 0);
  await stopOwnedProcess(proof, { exitNotification: closed });
});
