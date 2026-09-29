import assert from "node:assert/strict";
import { fork } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";
import test from "node:test";
import { acquireHostAdmission, processIdentityAlive, processIdentity } from "../../runtime/host-admission.mjs";
import { encodePNG, crc32 } from "../ui-review/png.mjs";
import { ReviewSession } from "../ui-review/session.mjs";
import { toolProfile } from "../ui-review/toolchain.mjs";
import { limits, parseRequest, schemaID, ReviewFailure } from "../ui-review/contract.mjs";
import { work } from "./ui-review-work-fixture.mjs";

const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const sessionAt = (root) => Object.assign(new ReviewSession({ UI_MODE: "artifacts" }, toolProfile()), { runtime: { privatePath: (...parts) => path.join(root, ...parts) } });
const caps = { cpu: 2, process: 2, memory_mb: 1536, io: 2, browser_stack: 2 };
const raster = { cpu: 1, process: 1, memory_mb: 1024, io: 1 }, report = { ...raster, memory_mb: 512 };

test("weighted siblings share inherited claims once; quiet waiters block newer activity but let existing work drain", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-weighted-")), leases = [];
  const acquire = async (options) => { const lease = await acquireHostAdmission({ root, capacities: caps, ...options }); leases.push(lease); return lease; };
  try {
    const parent = await acquire({ claims: raster });
    const first = await acquire({ parent: parent.token, claims: raster });
    const second = await acquire({ parent: parent.token, claims: report });
    await assert.rejects(acquire({ parent: parent.token, claims: report, timeoutMs: 150 }), /admission/u);
    const queue = [], quiet = acquire({ mode: "exclusive" }).then((lease) => { queue.push("quiet"); return lease; });
    await pause(150);
    const later = acquire({ claims: report }).then((lease) => { queue.push("later"); return lease; });
    await second.release(); await first.release();
    // Sub-work belongs to already admitted activity, even with a quiet waiter.
    const drain = await acquire({ parent: parent.token, claims: raster }); await drain.release();
    await parent.release(); const quietLease = await quiet;
    assert.deepEqual(queue, ["quiet"]); await quietLease.release(); await later;
    assert.deepEqual(queue, ["quiet", "later"]);
    for (const key of ["cpu", "process", "memory_mb", "io"]) await assert.rejects(acquire({ claims: { [key]: caps[key] + 1 } }), /admission/u);
  } finally { for (const lease of leases) await lease.release(); rmSync(root, { recursive: true, force: true }); }
});

test("worker deadlines and abort reap stalled computation before releasing staging and never publish a late result", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-stalled-work-"));
  try {
    const stall = path.join(root, "stall.mjs"); writeFileSync(stall, 'process.once("message", () => { while (true) {} });\n', { mode: 0o600 });
    for (const cancel of [false, true]) {
      const session = sessionAt(path.join(root, cancel ? "cancel" : "deadline"));
      const registration = Promise.withResolvers();
      let proof, acquired;
      const registered = new Promise((resolve) => { acquired = resolve; });
      const pending = work(session, "ui-review-report", "bundle-1", 1, { workerURL: pathToFileURL(stall).href, deadline: performance.now() + 5000,
        async onResource(resource) {
          if (resource.state !== "released") {
            proof = processIdentity(resource.target.pid); acquired();
            if (cancel) await registration.promise;
          }
        },
      });
      let timer;
      try {
        await Promise.race([registered, pending]);
        let ticks = 0; timer = setInterval(() => ticks++, 10);
        await pause(150);
        if (cancel) { session.abort.abort(new ReviewFailure("interrupted")); registration.resolve(); }
        await assert.rejects(pending, cancel ? /interrupted/u : /operation_expired/u);
        assert.ok(ticks > 0); assert.equal(processIdentityAlive(proof), false);
        assert.deepEqual(session.store.usage, { bytes: 0, bundles: 0, reservations: 0 });
      } finally {
        clearInterval(timer); session.abort.abort(new ReviewFailure("interrupted")); registration.resolve();
        await pending.catch(() => {}); session.store.close();
      }
    }
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test("worker proof retains weighted claims after parent death; nested quiet work excludes other roots", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-weighted-proof-")), leases = [];
  let child;
  const acquire = async (options) => { const lease = await acquireHostAdmission({ root, capacities: caps, ...options }); leases.push(lease); return lease; };
  try {
    const script = path.join(root, "child.mjs"); writeFileSync(script, 'process.send("ready"); setInterval(() => {}, 1000);\n', { mode: 0o600 });
    child = fork(script, [], { execArgv: [], stdio: ["ignore", "ignore", "ignore", "ipc"] });
    await new Promise((resolve) => child.once("message", resolve));
    const held = await acquire({ claims: raster }); await held.bind(processIdentity(child.pid));
    const file = path.join(root, "state.json"), state = JSON.parse(readFileSync(file));
    state.leases[0].process.start = "0"; writeFileSync(file, JSON.stringify(state));
    await assert.rejects(acquire({ claims: raster, timeoutMs: 150 }), /admission/u);
    const exited = new Promise((resolve) => child.once("exit", resolve)); child.kill("SIGKILL"); await exited; child = null;
    const parent = await acquire({ claims: raster }), other = await acquire({ claims: report });
    const quiet = acquire({ parent: parent.token, mode: "exclusive" }); await pause(150);
    await assert.rejects(acquire({ timeoutMs: 150 }), /admission/u);
    await other.release(); const quietLease = await quiet;
    await assert.rejects(acquire({ timeoutMs: 150 }), /admission/u);
    await quietLease.release(); const shared = await acquire({}); await shared.release();
    // Corrupt proof is never treated as evidence that an owner has died.
    const corrupt = JSON.parse(readFileSync(file)); corrupt.leases[0].process = { pid: process.pid };
    writeFileSync(file, JSON.stringify(corrupt)); await assert.rejects(acquire({ timeoutMs: 100 }), /admission/u);
    writeFileSync(file, JSON.stringify({ ...corrupt, leases: corrupt.leases.map((entry) => ({ ...entry, process: processIdentity() })) }));
  } finally { child?.kill("SIGKILL"); for (const lease of leases) await lease.release(); rmSync(root, { recursive: true, force: true }); }
});

test("maximum raster inputs and mixed parallel report jobs qualify declared worker memory envelopes", { timeout: 180000 }, async (t) => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-work-peaks-")), measurements = [];
  const left = sessionAt(path.join(root, "left")), right = sessionAt(path.join(root, "right"));
  const options = { onMeasurement: (value) => measurements.push(value) };
  try {
    // Both independent limits: maximum pixels/dimension and maximum PNG bytes.
    const raw = Buffer.alloc(limits.pixels * 4, 127);
    const png = await encodePNG({ width: limits.dimension, height: limits.pixels / limits.dimension, data: raw });
    const chunk = Buffer.alloc(limits.png - png.length); chunk.writeUInt32BE(chunk.length - 12); chunk.write("ruSt", 4); chunk.writeUInt32BE(crc32(chunk.subarray(4, -4)), chunk.length - 4);
    const file = path.join(root, "maximum.png"); writeFileSync(file, Buffer.concat([png.subarray(0, 33), chunk, png.subarray(33)]), { mode: 0o600 });
    await Promise.all([work(left, "ui-capture", { source: "image", path: file }, 1, options), work(right, "ui-capture", { source: "image", path: file }, 1, options)]);
    const request = parseRequest(Buffer.from(JSON.stringify({ schema_id: schemaID("analysis_request"), bundle_id: "bundle-1", operations: ["contact_sheet", "crop", "exact_diff"], comparison: { kind: "reference", bundle_id: "bundle-1" }, crops: Array.from({ length: 16 }, (_, index) => ({ x: 0, y: 0, width: limits.dimension - index, height: limits.pixels / limits.dimension })) })), "analysis_request");
    const [analyzed, rendered] = await Promise.all([work(left, "ui-analyze", request, 2, options), work(right, "ui-review-report", "bundle-1", 2, options)]);
    assert.equal(analyzed.bundle_id, "bundle-2"); assert.match(readFileSync(rendered.private_refs[0].absolute_path, "utf8"), /8192 × 2048/u);
    await work(left, "ui-review-report", "bundle-2", 3, options);
    for (const measurement of measurements) { assert.ok(measurement.peak_bytes > 0); assert.ok(measurement.peak_bytes <= measurement.limit_bytes); assert.equal(measurement.reaped, true); }
    for (const kind of ["raster", "report"]) t.diagnostic(JSON.stringify({ kind, measured_peak_bytes: Math.max(...measurements.filter((entry) => entry.kind === kind).map((entry) => entry.peak_bytes)), declared_bytes: (kind === "raster" ? 1024 : 512) * 1048576, platform: process.platform, architecture: process.arch, kernel: os.release() }));
  } finally { left.store.close(); right.store.close(); rmSync(root, { recursive: true, force: true }); }
});
