import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import { createServer } from "node:net";
import { syncBuiltinESMExports } from "node:module";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { createSuiteRuntime } from "../runtime/suite-runtime.mjs";

const scenarios = ["allocation", "status", "second-open", "spawn", "spawn-sync", "tail-read", "read-and-cleanup", "cleanup-retry", "close", "parent-mode", "parent-symlink", "parent-owner", "file-mode", "file-symlink", "tail-limits", "component-length", "component-character", "cancellation", "snapshot-caller", "make-wrapper"];

// Fault injection lives in isolated subprocesses; importing this case module is inert.
export function assertPrivateChildCaptureBoundary() {
  for (const scenario of scenarios) {
    assert.equal(execFileSync(process.execPath, [fileURLToPath(import.meta.url), scenario], {
      encoding: "utf8", timeout: 30_000,
    }), `${scenario}: pass\n`, scenario);
  }
}

async function exercise(scenario) {
  const fixture = fs.mkdtempSync(path.join(tmpdir(), "cartulary-capture-contract-"));
  const repoRoot = path.join(fixture, "repo");
  const runRoot = path.join(fixture, "results", "capture-contract");
  fs.mkdirSync(repoRoot, { mode: 0o700 });
  fs.mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  const runtime = createSuiteRuntime({ repoRoot, runRoot, runID: "capture-contract", scratchRoot: path.join(fixture, "scratch") });
  const parent = runtime.privatePath("child-captures");
  fs.writeFileSync(path.join(fixture, "protected"), "protected bytes", { mode: 0o600 });
  const options = { repoRoot, runRoot, cwd: repoRoot, env: {
    ...process.env,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runtime.runID,
  } };
  const originals = new Map();
  const patch = (name, implementation) => {
    if (!originals.has(name)) originals.set(name, fs[name]);
    fs[name] = implementation;
  };
  const restore = () => {
    for (const [name, implementation] of originals) fs[name] = implementation;
    syncBuiltinESMExports();
  };
  const descriptors = new Map();
  const originalOpen = fs.openSync;
  const originalClose = fs.closeSync;
  const originalRead = fs.readSync;
  const originalUnlink = fs.unlinkSync;
  const originalStat = fs.lstatSync;
  let injected = false;
  let cleanupInjected = false;
  patch("openSync", (file, flags, ...rest) => {
    if (scenario === "second-open" && String(file).startsWith(parent) && path.basename(file) === "stderr" && flags & fs.constants.O_CREAT) {
      injected = true;
      throw new Error("injected second stream open");
    }
    const descriptor = originalOpen(file, flags, ...rest);
    if (String(file).startsWith(`${parent}/`)) descriptors.set(descriptor, String(file));
    return descriptor;
  });
  patch("closeSync", (descriptor) => {
    if (scenario === "close" && descriptors.has(descriptor) && !injected) {
      injected = true;
      throw new Error("injected descriptor close");
    }
    originalClose(descriptor);
    descriptors.delete(descriptor);
  });
  patch("readSync", (descriptor, ...args) => {
    if (["tail-read", "read-and-cleanup"].includes(scenario) && descriptors.get(descriptor)?.endsWith("/stderr")) {
      injected = true;
      throw new Error("injected tail read");
    }
    return originalRead(descriptor, ...args);
  });
  patch("unlinkSync", (file) => {
    if (["read-and-cleanup", "cleanup-retry"].includes(scenario) && String(file).startsWith(parent) && path.basename(file) === "stdout" && !cleanupInjected) {
      cleanupInjected = true;
      throw new Error("injected unlink");
    }
    return originalUnlink(file);
  });
  if (["component-length", "component-character"].includes(scenario)) patch("mkdtempSync", (prefix) => {
    const directory = `${prefix}${scenario === "component-length" ? "x".repeat(128) : "unsafe name"}`;
    fs.mkdirSync(directory, { mode: 0o700 });
    return directory;
  });
  if (scenario === "parent-mode") fs.mkdirSync(parent, { mode: 0o755 });
  if (scenario === "parent-symlink") fs.symlinkSync(repoRoot, parent);
  if (scenario === "parent-owner") patch("lstatSync", (file, ...args) => {
    const stat = originalStat(file, ...args);
    if (file === parent) stat.uid = process.getuid() + 1;
    return stat;
  });
  if (["file-mode", "file-symlink"].includes(scenario)) patch("lstatSync", (file, ...args) => {
    if (String(file).startsWith(parent) && path.basename(file) === "stdout" && !injected) {
      injected = true;
      if (scenario === "file-mode") fs.chmodSync(file, 0o644);
      else { originalUnlink(file); fs.symlinkSync(path.join(fixture, "protected"), file); }
    }
    return originalStat(file, ...args);
  });
  syncBuiltinESMExports();
  try {
    const { runPrivateCapturedProcess: capture } = await import("../runtime/private-child-process.mjs");
    const run = (source = 'process.stdout.write("output"); process.stderr.write("diagnostic");', extra = {}) =>
      capture(process.execPath, ["--eval", source], { ...options, ...extra });
    if (scenario === "allocation") {
      const catalog = JSON.parse(fs.readFileSync(new URL("../../test_families/module.timeline.json", import.meta.url), "utf8"));
      const rows = catalog.rows.filter((row) => row.selector?.file === "apps/web/e2e/measurement/timeline-grid.spec.ts");
      assert.equal(rows.length, 4);
      const labels = rows.map((row) => row.row_id);
      const results = await Promise.all([...labels, ...labels].map((label) => run("process.stdout.write(process.env.SEMANTIC_LABEL)", { env: { ...options.env, SEMANTIC_LABEL: label } })));
      assert.equal(new Set(results.map((result) => result.stdoutPath)).size, 8);
      for (const [index, result] of results.entries()) {
        assert.equal(result.stdout, labels[index % 4]);
        assert.equal(result.status, 0);
        assert.equal(result.signal, null);
        const directory = path.dirname(result.stdoutPath);
        assert.match(path.basename(directory), /^capture-[A-Za-z0-9_.-]+$/u);
        assert.ok(path.basename(directory).length <= 128);
        assert.equal(fs.lstatSync(directory).mode & 0o777, 0o700);
        for (const file of [result.stdoutPath, result.stderrPath]) assert.equal(fs.lstatSync(file).mode & 0o777, 0o600);
        result.cleanup();
        result.cleanup();
        assert.ok(!fs.existsSync(directory));
      }
      const again = await run();
      assert.ok(!results.some((result) => result.stdoutPath === again.stdoutPath));
      again.cleanup();
    } else if (scenario === "status") {
      const result = await run('process.stdout.write("x".repeat(8192)); process.stderr.write("y".repeat(8192)); process.exitCode = 7;', { tailBytes: 1024 });
      assert.equal(result.status, 7);
      assert.equal(result.stdout, "x".repeat(1024));
      assert.equal(result.stderr, "y".repeat(1024));
      result.cleanup();
      const signal = await run('process.kill(process.pid, "SIGTERM")');
      assert.equal(signal.status, null);
      assert.equal(signal.signal, "SIGTERM");
      signal.cleanup();
    } else if (scenario === "cleanup-retry") {
      const result = await run();
      assert.throws(() => result.cleanup(), { failure_reason: "cleanup_error" });
      assert.ok(!fs.existsSync(result.stderrPath), "remaining release attempted after first unlink failure");
      assert.ok(fs.existsSync(result.stdoutPath));
      result.cleanup();
      result.cleanup();
    } else if (["snapshot-caller", "make-wrapper"].includes(scenario)) {
      const { executeUnitProcess } = await import("../scheduler/work-graph/executor.mjs");
      const root = path.resolve(import.meta.dirname, "../../..");
      const preload = path.join(fixture, "capture-fault.mjs");
      fs.writeFileSync(preload, `
        import fs from "node:fs";
        import { syncBuiltinESMExports } from "node:module";
        const unlink = fs.unlinkSync;
        const open = fs.openSync;
        const mkdtemp = fs.mkdtempSync;
        fs.mkdtempSync = (prefix, ...args) => {
          const directory = mkdtemp(prefix, ...args);
          if (String(prefix).includes("/child-captures/")) fs.writeFileSync(${JSON.stringify(path.join(fixture, "capture-allocation.json"))}, JSON.stringify(directory), { mode: 0o600 });
          return directory;
        };
        const capturedOutput = (file) => String(file).includes("/child-captures/") && String(file).endsWith("/stdout");
        fs.unlinkSync = (file) => {
          if (capturedOutput(file)) throw new Error("injected caller cleanup failure");
          return unlink(file);
        };
        if (${JSON.stringify(scenario)} === "make-wrapper") fs.openSync = (file, flags, ...args) => {
          if (capturedOutput(file) && (flags & fs.constants.O_CREAT) === 0) throw new Error("injected caller read failure");
          return open(file, flags, ...args);
        };
        syncBuiltinESMExports();
      `, { mode: 0o600 });
      const environment = { ...process.env };
      for (const key of Object.keys(environment)) {
        if (key.startsWith("CARTULARY_") || key === "TASK_SURFACE_MANIFEST") delete environment[key];
      }
      Object.assign(environment, {
        NODE_OPTIONS: `--import=${preload}`,
        CARTULARY_TEST_RESULTS_DIR: path.dirname(runRoot),
        CARTULARY_TEST_RUN_ID: runtime.runID,
        CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root,
        CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
        CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runtime.runID,
      });
      const execute = (args, extra = {}) => executeUnitProcess({ unit_id: "capture-caller", kind: "fixture_builder", timeout_ms: 10_000,
        command: { executable: process.execPath, args, environment: {} } },
      { cwd: root, environment: { ...environment, ...extra }, inheritProcessEnvironment: false });
      if (scenario === "snapshot-caller") {
        const { loadPerformanceFixtureSnapshotRegistry, postgresMigrationDigest, snapshotKey } = await import("../performance-fixture/index.mjs");
        const profile = loadPerformanceFixtureSnapshotRegistry(root).profiles.get("ac043_large_grid_snapshot_v1");
        const key = snapshotKey(profile, postgresMigrationDigest(root));
        const executable = path.join(fixture, "fixture-builder");
        for (const status of [0, 7]) {
          fs.writeFileSync(executable, `#!${process.execPath}\nif (${status} !== 0) process.stderr.write("fixture primary failure\\n"); process.exitCode = ${status};\n`, { mode: 0o700 });
          const outcome = await execute(["tools/harness/performance-fixture/snapshot-builder-cli.mjs", "--fixture-profile", profile.fixture_profile_id, "--snapshot-key", key], {
            CARTULARY_TEST_SERVICES_BIN: executable,
            CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID: "fixture-builder-contract",
          });
          assert.equal(outcome.exit_code, status === 0 ? 12 : 1, JSON.stringify(outcome));
          assert.ok(outcome.stderr.includes("cleanup_error"), "secondary cleanup failure stays visible");
          if (status === 0) assert.equal(outcome.failure_reason, "cleanup_error", "scheduler preserves cleanup taxonomy without a command channel");
          else {
            assert.notEqual(outcome.failure_reason, "cleanup_error");
            assert.ok(outcome.stderr.includes("fixture primary failure"));
          }
          assert.ok(!outcome.stderr.includes(runtime.root));
        }
        for (const directory of fs.readdirSync(parent)) assert.deepEqual(fs.readdirSync(path.join(parent, directory)), ["stdout"]);
      } else {
        const wrapperResults = path.join(fixture, "wrapper-results");
        const target = "go-test-duration-baseline-drift";
        const outcome = await execute(["tools/harness/execution/run-make-node-tool-cli.mjs", target], {
          CARTULARY_TEST_RESULTS_DIR: wrapperResults,
          CARTULARY_TEST_RUN_ID: "wrapper-run",
          CARTULARY_OUTPUT_MODE: "summary",
          RESULTS_DIR: path.join(fixture, "missing-evidence"),
        });
        assert.equal(outcome.exit_code, 11, JSON.stringify(outcome));
        const targetRoot = path.join(wrapperResults, "wrapper-run", target);
        const summary = JSON.parse(fs.readFileSync(path.join(targetRoot, "tool-run-summary.json"), "utf8"));
        assert.equal(summary.failure_reason, "artifact_error");
        const diagnostic = fs.readFileSync(path.join(targetRoot, "stderr.log"), "utf8");
        assert.match(diagnostic, /cleanup_error/u, "wrapper must retain the helper's bounded secondary failure");
        assert.ok(!diagnostic.includes(fixture));
        const allocation = JSON.parse(fs.readFileSync(path.join(fixture, "capture-allocation.json"), "utf8"));
        assert.ok(!fs.existsSync(allocation), "suite cleanup removes the actual failed invocation allocation");
      }
    } else if (scenario === "cancellation") {
      const { executeUnitProcess } = await import("../scheduler/work-graph/executor.mjs");
      const controller = new AbortController();
      let childPID;
      const server = createServer((socket) => {
        socket.once("data", (data) => { childPID = Number(data.toString()); controller.abort(); });
      });
      await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
      const source = path.join(fixture, "capture-consumer.mjs");
      // The consumer handles TERM only to await/reap its child; suite cleanup
      // remains responsible for private material after the interrupted consumer.
      fs.writeFileSync(source, `
        import { runPrivateCapturedProcess } from ${JSON.stringify(new URL("../runtime/private-child-process.mjs", import.meta.url).href)};
        process.on("SIGTERM", () => {});
        await runPrivateCapturedProcess(process.execPath, ["--input-type=module", "--eval", ${JSON.stringify(`
          import { connect } from "node:net";
          const socket = connect(${server.address().port}, "127.0.0.1", () => socket.write(String(process.pid)));
          process.stdout.write("private cancellation output");
        `)}], ${JSON.stringify(options)});
      `, { mode: 0o600 });
      try {
        const outcome = await executeUnitProcess({ unit_id: "capture-cancellation", kind: "runner", timeout_ms: 10_000,
          command: { executable: process.execPath, args: [source], environment: {} } },
        { cwd: repoRoot, environment: options.env, signal: controller.signal });
        assert.equal(outcome.failure_reason, "cancelled_or_interrupted");
        assert.ok(Number.isInteger(childPID) && childPID > 0, "cancel only after captured child readiness");
        assert.throws(() => process.kill(childPID, 0), { code: "ESRCH" }, "child reaped before suite cleanup");
        assert.equal(outcome.stdout, "");
        assert.equal(outcome.stderr, "");
        assert.equal(fs.readdirSync(parent).length, 1, "suite owns interrupted consumer residue");
      } finally { await new Promise((resolve) => server.close(resolve)); }
    } else if (scenario === "tail-limits") {
      for (const tailBytes of [0, 1023, 1024 * 1024 + 1, Infinity]) await assert.rejects(run(undefined, { tailBytes }), { failure_reason: "artifact_error" });
      const result = await run(undefined, { tailBytes: 1024 * 1024 });
      result.cleanup();
    } else {
      const invocation = scenario === "spawn" ? capture(path.join(fixture, "absent"), [], options)
        : scenario === "spawn-sync" ? capture(null, [], options) : run();
      await assert.rejects(invocation, (error) => {
        assert.equal(error.failure_reason, "artifact_error");
        assert.ok(!error.message.includes(runtime.root));
        if (scenario === "read-and-cleanup") assert.ok(error.cause.errors.some((cause) => cause.failure_reason === "cleanup_error"));
        return true;
      });
      if (scenario === "parent-mode") assert.equal(fs.lstatSync(parent).mode & 0o777, 0o755, "unsafe mode must not be repaired");
      if (["second-open", "tail-read", "read-and-cleanup", "close", "file-mode", "file-symlink"].includes(scenario)) assert.ok(injected);
    }
    assert.equal(descriptors.size, 0, "all acquired descriptors closed");
    if (fs.existsSync(parent) && !["parent-symlink", "read-and-cleanup", "cancellation", "snapshot-caller"].includes(scenario)) assert.deepEqual(fs.readdirSync(parent), []);
    if (scenario === "read-and-cleanup") {
      const [directory] = fs.readdirSync(parent);
      assert.deepEqual(fs.readdirSync(path.join(parent, directory)), ["stdout"], "only the failed release remains for suite cleanup");
    }
    assert.equal(fs.readFileSync(path.join(fixture, "protected"), "utf8"), "protected bytes");
    assert.ok(fs.existsSync(runtime.root), "capture must not remove borrowed suite root");
    assert.deepEqual(fs.readdirSync(runRoot), [], "raw output and private paths are never retained");
  } finally {
    restore();
    runtime.close();
    assert.ok(!fs.existsSync(runtime.root));
    fs.rmSync(fixture, { recursive: true, force: true });
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const scenario = process.argv[2];
  assert.ok(scenarios.includes(scenario));
  await exercise(scenario);
  process.stdout.write(`${scenario}: pass\n`);
}
