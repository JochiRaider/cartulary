import assert from "node:assert/strict";
import test from "node:test";
import { EventEmitter } from "node:events";
import { mkdtempSync, rmSync } from "node:fs";
import path from "node:path";
import os from "node:os";
import { createCollectionEngine } from "../collection-engine.mjs";
import { createDiagnosticSession } from "../diagnostic-session.mjs";
import { createResourceCollector, instrumentationPolicy } from "../resource-collector.mjs";
import { createResourceStore } from "../resource-store.mjs";

const deferred = () => { let resolve, reject; const promise = new Promise((yes, no) => { resolve = yes; reject = no; }); return { promise, resolve, reject }; };
const settle = async () => { for (let i = 0; i < 30; i += 1) await Promise.resolve(); };
const digest = `sha256:${"a".repeat(64)}`;
const identity = { run_id: "engine-fixture", source_digest: digest, graph_digest: digest, policy_digest: digest };
const proof = { identity: digest, pid: 1, start: "1", stat: { parent: 0 } };
const sample = { availability: "available", metrics: { rss_bytes: { value: 12, availability: "available" } } };
function fixture(overrides = {}) {
  const timers = new Map(), records = [], statuses = [], calls = [];
  let tick = 0, next = 0, index;
  const clock = { now: () => tick, elapsed: () => tick, setTimeout: (fn) => { timers.set(++next, fn); return next; }, clearTimeout: (id) => timers.delete(id) };
  const gate = { quiet: () => false, acquire: async () => ({ release: async () => calls.push("released") }), close: () => calls.push("gate_closed"), ...overrides.gate };
  const store = { samples: 0, bytes: 0, truncated: false, append: async (value) => { records.push(value); store.samples += 1; }, finish: async (value) => { index = value; calls.push("published"); }, ...overrides.store };
  const adapter = { discover: async () => ({ found: [], truncated: false }), processSample: async () => sample, context: async () => [], bytesRead: 0, ...overrides.adapter };
  const engine = createCollectionEngine({ policy: { ...instrumentationPolicy().policy, ...overrides.policy }, identity, gate, store, adapter, clock,
    observer: { heap: () => 10, cpu: () => ({ user: 2, system: 3 }) }, onStatus: (status) => statuses.push(status) });
  engine.register(proof, "unit:a");
  return { engine, gate, store, records, statuses, calls, timers, get index() { return index; }, async fire() { tick += 2000; const [id, fn] = timers.entries().next().value; timers.delete(id); fn(); await settle(); } };
}

test("pause cancels pending admission immediately; stop dominates resume and shares completion", async () => {
  const admission = deferred(), released = deferred(); let signal;
  const f = fixture({ gate: { acquire: (options) => { signal = options.signal; return admission.promise; } } });
  await f.fire();
  let paused = false;
  const pause = f.engine.pause().then(() => { paused = true; });
  assert.equal(signal.aborted, true);
  const stop = f.engine.stop();
  assert.equal(f.engine.stop(), stop);
  assert.equal(f.engine.resume(), stop);
  admission.resolve({ release: () => released.promise });
  await settle(); assert.equal(paused, false); assert.equal(f.index, undefined);
  released.resolve(); await pause; await stop;
  assert.equal(f.records.length, 0); assert.equal(f.timers.size, 0);
  assert.equal(f.index.failed_sweeps, 0);
  assert.deepEqual(f.calls, ["gate_closed", "published"]);
});

for (const operation of ["read", "write"]) test(`pause drains an in-flight ${operation} and its fence; resume starts a new segment`, async () => {
  const pending = deferred(), release = deferred();
  const f = fixture({ adapter: operation === "read" ? { processSample: () => pending.promise } : {},
    store: operation === "write" ? { append: () => pending.promise } : {},
    gate: { acquire: async () => ({ release: () => release.promise }) } });
  await f.fire(); let done = false;
  const pause = f.engine.pause().then(() => { done = true; });
  await settle(); assert.equal(done, false);
  pending.resolve(sample); await settle(); assert.equal(done, false);
  release.resolve(); await pause; assert.equal(f.timers.size, 0);
  f.gate.acquire = async () => ({ release: async () => {} });
  f.store.append = async (record) => f.records.push(record);
  await f.engine.resume(); await f.fire();
  assert.ok(f.records.every((record) => record.segment > 0));
  await f.engine.stop();
});

test("watcher failure is disclosed and leaves collection quiescent until explicit resume", async () => {
  let failed;
  const f = fixture({ gate: { quiet: () => true, waitUntilAvailable: (_ready, error) => { failed = error; return () => {}; } } });
  await f.fire(); failed(); await settle();
  assert.equal(f.timers.size, 0); assert.equal(f.statuses.at(-1).availability, "collector_failed");
  await f.engine.stop(); assert.equal(f.index.failed_sweeps, 1); assert.equal(f.index.status, "partial");
});

test("read failures, discovery bounds and lifetime caps remain independent of terminal lifecycle", async () => {
  const f = fixture({ adapter: { discover: async () => ({ found: [], truncated: true }), processSample: async () => { throw new Error("read"); } }, policy: { maximum_lifetime_processes: 1 } });
  await f.fire();
  assert.equal(f.records[0].availability, "unsupported");
  f.engine.register({ ...proof, identity: "other" }, null);
  await f.engine.stop();
  assert.equal(f.index.processes.length, 1); assert.equal(f.index.status, "truncated");
  assert.equal(f.index.discovery_truncations, 1); assert.equal(f.index.omitted_observations, 1);
});

test("write rejection is retained as diagnostic failure and publication rejection reaches all stop callers", async () => {
  const f = fixture({ store: { append: async () => { throw new Error("ENOSPC"); }, finish: async () => { throw new Error("interrupted publication"); } } });
  await f.fire(); assert.equal(f.statuses.at(-1).failed_sweeps, 1);
  const a = f.engine.stop(), b = f.engine.stop(); assert.equal(a, b);
  await assert.rejects(a, /interrupted publication/u); await assert.rejects(b, /interrupted publication/u);
});

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

test("store caps and disk-full failures retain honest metadata with a single writer", async (t) => {
  const runRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-store-")); t.after(() => rmSync(runRoot, { recursive: true, force: true }));
  for (const fail of [false, true]) {
    let index, aborted = false, writes = 0;
    const store = createResourceStore({ runRoot, policy: { ...instrumentationPolicy().policy, maximum_sample_bytes: fail ? 9999 : 1 },
      createWriter: () => ({ write: async () => { writes += 1; throw new Error("ENOSPC"); }, close: async () => {}, abort: async () => { aborted = true; } }),
      publish: (_file, data) => { index = JSON.parse(data); } });
    const record = { schema_id: "cartulary.harness_resource_sample.v1", seq: 1, elapsed_ms: 0, segment: 0, identity_digest: digest, scope: "process", scope_ref: "process:1", ...sample };
    if (fail) await assert.rejects(store.append(record), /ENOSPC/u); else await store.append(record);
    const f = fixture(); await f.engine.stop();
    await store.finish(f.index, 0, () => 1);
    assert.equal(index.status, fail ? "collector_failed" : "truncated"); assert.equal(index.samples, 0);
    assert.equal(writes, fail ? 1 : 0); assert.equal(aborted, fail);
    assert.equal(index.sample_digest === null, fail);
  }
});

test("a failed fence release cannot acknowledge pause or permit another sweep", async () => {
  const released = deferred();
  const f = fixture({ gate: { acquire: async () => ({ release: () => released.promise }) } });
  await f.fire();
  const pause = f.engine.pause();
  released.reject(new Error("release failed"));
  await assert.rejects(pause, /release failed/u);
  await assert.rejects(f.engine.stop(), /release failed/u);
  assert.equal(f.timers.size, 0); assert.equal(f.index, undefined);
});
