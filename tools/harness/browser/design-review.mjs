import { spawn, execFileSync } from "node:child_process";
import { createHash, randomUUID } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { FixtureBroker, productionFixtureProviders, startManagedSuite } from "../scheduler/fixture-broker/index.mjs";
import { createSuiteRuntime, scanRetainedRoot } from "../runtime/suite-runtime.mjs";
import { buildSourceSnapshot } from "../test-catalog/index.mjs";
import { seedDesignReview } from "./design-review-seed.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const target = process.argv.includes("--smoke") ? "browser-design-review-smoke" : "browser-design-review";

export function reviewProfile(value = "network_flow_claimed") {
  value = value.trim();
  if (!["default", "network_flow_claimed"].includes(value)) {
    throw new Error("REVIEW_PROFILE must be default or network_flow_claimed");
  }
  return value;
}

function json(file, value) {
  mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`, { mode: 0o600 });
}

function child(command, args, environment, signal, output = "inherit") {
  return new Promise((resolve, reject) => {
    signal?.throwIfAborted();
    const subprocess = spawn(command, args, { cwd: root, env: environment, stdio: output, detached: true });
    const stop = () => {
      try { process.kill(-subprocess.pid, "SIGTERM"); } catch (error) { if (error.code !== "ESRCH") reject(error); }
    };
    signal?.addEventListener("abort", stop, { once: true });
    subprocess.once("error", (error) => { signal?.removeEventListener("abort", stop); reject(error); });
    subprocess.once("close", (code, reason) => {
      signal?.removeEventListener("abort", stop);
      if (code === 0 && !signal?.aborted) resolve();
      else reject(new Error(`${path.basename(command)} failed (${reason ?? code})`));
    });
  });
}

/** Keep cleanup ordered and attempt every owner even after a failed finalizer. */
export async function withReviewResources({ acquire, prepare, hold, close, finish }) {
  let primary;
  try {
    const resource = await acquire();
    await prepare(resource);
    await hold(resource);
  } catch (error) {
    primary = error;
  } finally {
    for (const cleanup of [close, finish]) {
      try { await cleanup(primary); } catch (error) { primary ??= error; }
    }
  }
  if (primary) throw primary;
}

export async function holdReviewSession({ check, signal, intervalMs = 3000 }) {
  while (!signal?.aborted) {
    try { await check(); } catch (error) { if (signal?.aborted) return; throw error; }
    if (signal?.aborted) return;
    await new Promise((resolve) => {
      const done = () => { clearTimeout(timer); signal?.removeEventListener("abort", done); resolve(); };
      const timer = setTimeout(done, intervalMs);
      signal?.addEventListener("abort", done, { once: true });
    });
  }
}

export async function runDesignReview({ environment = process.env, signal, onReady, hold, verifySamples = false } = {}) {
  const profile = reviewProfile(environment.REVIEW_PROFILE);
  const runID = `design-review-${Date.now()}-${randomUUID().slice(0, 8)}`;
  const runRoot = path.join(root, ".cartulary/test-results", runID);
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  // The lifecycle requires a source descriptor. This is deliberately not a
  // harness_run_manifest or a claim that catalog rows executed. Resolve source
  // inputs before acquiring private runtime resources.
  json(path.join(runRoot, "run-manifest.json"), {
    purpose: "interactive_design_review",
    source_digest: buildSourceSnapshot(root).digest,
    toolchain_digest: `sha256:${createHash("sha256").update(readFileSync(path.join(root, "tools/toolchain_pins.json"))).digest("hex")}`,
    source_commit: execFileSync("git", ["rev-parse", "HEAD"], { cwd: root, encoding: "utf8" }).trim(),
    runtime_profile_id: profile,
  });
  const runtime = createSuiteRuntime({ repoRoot: root, runRoot, runID });
  const base = { ...environment };
  for (const name of Object.keys(base)) {
    if (/^(?:MAKEFLAGS|MAKEOVERRIDES|MFLAGS|REVIEW_PROFILE|CARTULARY_(?:TEST_|WEB_E2E_|BROWSER_|HARNESS_|PGTEST_|S3_)|CARTULARY__)/u.test(name)) delete base[name];
  }
  Object.assign(base, {
    CARTULARY_TEST_RESULTS_DIR: path.dirname(runRoot),
    CARTULARY_TEST_RUN_ID: runID,
    CARTULARY_TEST_TARGET: target,
    CARTULARY_HARNESS_IDENTITY_PREPARED: "1",
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID,
    CARTULARY_SERVER_HARNESS_BIN: path.join(root, "server-harness"),
    CARTULARY_MIGRATE_BIN: path.join(root, "migrate"),
    CARTULARY_TEST_SERVICES_BIN: path.join(root, "tmp/toolbin/cartulary-test-services"),
    NODE_RUNTIME_DIR: path.join(root, "tmp/node-runtime"),
    NODE_BIN: process.execPath,
    PNPM: path.join(root, "tmp/node-runtime/bin/pnpm"),
    CARTULARY_BROWSER_RUNTIME_PROFILE_ID: profile,
    CARTULARY_BROWSER_SERVICE_REQUIREMENT: "test-services",
    CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED: "false",
    CARTULARY__ENTERPRISE_AUTHENTICATION__CLAIMED: "false",
  });
  let suite;
  let broker;
  let state = "preparing";
  const suiteController = {
    ensure() {
      suite ??= startManagedSuite({ root, target, suiteRuntime: runtime, environment: base });
      return suite;
    },
    close() { suite?.close(); },
  };
  let sessionDetails = {};
  const terminal = (extra = {}) => {
    sessionDetails = { ...sessionDetails, ...extra };
    json(path.join(runRoot, "review-session.json"), {
      purpose: "interactive_design_review", run_id: runID, runtime_profile_id: profile, state, ...sessionDetails,
    });
  };
  terminal();
  process.stdout.write(`Preparing browser review. Diagnostics: ${runRoot}\n`);
  await withReviewResources({
    async acquire() {
      const buildEnvironment = { ...base, CARTULARY_HARNESS_GRAPH_CHILD: "1", CARTULARY_HARNESS_GRAPH_ARTIFACT_CHILD: "1" };
      for (const prerequisite of ["build-web", "build-server-harness", "build-migrate", "test-service-images", "playwright-install"]) {
        await child("make", ["--no-print-directory", prerequisite], { ...buildEnvironment, CARTULARY_TEST_TARGET: prerequisite }, signal);
      }
      signal?.throwIfAborted();
      broker = new FixtureBroker({
        providers: productionFixtureProviders({ root, runtimeEnvironment: base, suiteRuntime: runtime, suiteController }),
        recordSink: (record) => json(path.join(runRoot, "_shared/fixture-leases", `${record.lease_id}.json`), record),
      });
      const lease = await broker.acquire("browser_stack", { affinityKey: `design-review-${profile}`, unitID: target, browserStage: "webserver-backed", runtimeProfileID: profile });
      signal?.throwIfAborted();
      return { ...base, ...lease.resource.environment };
    },
    async prepare(attached) {
      state = "seeding";
      terminal();
      // Use exactly the attach-only guard used by retained browser execution.
      const assignment = execFileSync(process.execPath, ["tools/harness/browser/browser-session-evidence.mjs", "attach-json", attached.CARTULARY_WEB_E2E_STACK_JSON_FILE], { cwd: root, env: attached, encoding: "utf8" });
      Object.assign(attached, JSON.parse(assignment));
      const privateDirectory = runtime.privatePath("review-access");
      mkdirSync(privateDirectory, { recursive: true, mode: 0o700 });
      const review = await seedDesignReview({ root, environment: attached, privateDirectory, runRoot, profile, signal, verifySamples, registerSecret: (value) => runtime.registerSecret(value) });
      state = "ready";
      terminal({ public_origin: attached.CARTULARY_WEB_E2E_PUBLIC_ORIGIN, scenarios: review.scenarios, samples: review.samples });
      process.stdout.write(`Browser review ready: ${attached.CARTULARY_WEB_E2E_PUBLIC_ORIGIN}\nPrivate login instructions: ${path.join(privateDirectory, "access.json")}\nSample files: ${review.samples}\nScenario index: ${path.join(runRoot, "review-session.json")}\n`);
      for (const scenario of review.scenarios) process.stdout.write(`  ${scenario.name}: ${scenario.url}\n`);
      process.stdout.write(hold ? "Smoke complete; cleaning up owned resources.\n" : "Press Ctrl-C to stop and remove this session's data.\n");
      await onReady?.({ attached, review, runRoot, privateDirectory });
    },
    async hold(attached) {
      if (hold) return hold(attached);
      // Check process identity, immutable build, and active ownership without
      // blocking signal handling or misclassifying interruption as a dead stack.
      await holdReviewSession({ signal, check: () => child(process.execPath,
        ["tools/harness/browser/browser-session-evidence.mjs", "attach-json", attached.CARTULARY_WEB_E2E_STACK_JSON_FILE], attached, signal, "ignore") });
    },
    async close(error) {
      state = error ? "failed" : "closed";
      try { await broker?.close(); } finally { suiteController.close(); }
    },
    async finish(error) {
      if (error) state = "failed";
      terminal({ ...(error ? { failure: error.message } : {}) });
      try {
        const scan = await scanRetainedRoot(runRoot, { forbiddenValues: runtime.forbiddenValues(), removeUnsafe: true });
        json(path.join(runRoot, "retained-secret-scan.json"), scan);
        if (scan.status !== "pass") throw new Error("Review retained-artifact secret scan failed");
      } catch (scanError) {
        state = "failed";
        terminal({ artifact_failure: scanError.message });
        throw scanError;
      } finally { runtime.close(); }
    },
  });
  return runRoot;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const controller = new AbortController();
  const stop = () => controller.abort();
  process.on("SIGINT", stop);
  process.on("SIGTERM", stop);
  try {
    if (process.argv.slice(2).some((arg) => arg !== "--smoke")) throw new Error("usage: design-review.mjs [--smoke]");
    await runDesignReview({ signal: controller.signal, ...(process.argv.includes("--smoke") ? { hold: async () => {}, verifySamples: true } : {}) });
  }
  catch (error) {
    process.stderr.write(`Browser review failed: ${error.message}\n`);
    process.exitCode = controller.signal.aborted ? 130 : 1;
  } finally {
    process.off("SIGINT", stop);
    process.off("SIGTERM", stop);
  }
}
