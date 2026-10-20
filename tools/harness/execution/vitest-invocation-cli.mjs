#!/usr/bin/env node
import { randomUUID } from "node:crypto";
import { createRequire } from "node:module";
import { mkdirSync, rmSync } from "node:fs";
import { Writable } from "node:stream";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { primaryPublicFailure, publicExitCodeForFailure, publicExitCodeForFailures, redactString } from "../contract/index.mjs";
import { createVitestEvidence, diagnosticFailure, readVitestJSON } from "../diagnostics/vitest-failure-details.mjs";
import VitestCollector from "../diagnostics/vitest-collector.mjs";
import { adaptVitestInvocation, validateVitestRequest, vitestInvocation } from "./runners/vitest.mjs";
import { publishRowResult } from "./runners/row-results.mjs";
import { settledForkPool } from "./runners/vitest-worker-lifecycle.mjs";
import { reportCommandFailure } from "../runtime/command-failure.mjs";
import { borrowRuntime } from "../runtime/session-client.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
function regexEscape(value) { return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&"); }

// Vitest pipes each fork's stderr with end=true. The destination belongs to the
// invocation, so a worker's EOF must not close the coordinator's diagnostic sink.
class InvocationStderr extends Writable {
  end(chunk, encoding, callback) {
    if (typeof chunk === "function") { callback = chunk; chunk = undefined; }
    else if (typeof encoding === "function") { callback = encoding; encoding = undefined; }
    if (chunk !== undefined && chunk !== null) this.write(chunk, encoding, callback);
    else callback?.();
    return this;
  }
  close() { super.end(); }
}

async function main() {
  const [reportFile, detailsFile, requestFile, ...extra] = process.argv.slice(2);
  if (!reportFile || !detailsFile || !requestFile || extra.length) throw diagnosticFailure("Vitest invocation requires report, diagnostics, and exact request paths");
  const request = readVitestJSON(requestFile);
  validateVitestRequest(request);
  const invocation = vitestInvocation(root, request);
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  const runRoot = path.resolve(root, process.env.CARTULARY_TEST_RESULTS_DIR, runID);
  for (const file of [reportFile, detailsFile, requestFile]) {
    const relative = path.relative(runRoot, path.resolve(file));
    if (relative.startsWith("..") || path.isAbsolute(relative)) throw diagnosticFailure("Vitest artifact escapes its run root");
  }
  const runtime = borrowRuntime({ repoRoot: root, runRoot });
  const invocationID = randomUUID();
  const privateRoot = runtime.privatePath(`vitest-${invocationID}`);
  mkdirSync(privateRoot, { mode: 0o700 });
  const rawReport = path.join(privateRoot, "runner.json");
  const startedAt = new Date().toISOString(), started = Date.now();
  const collector = new VitestCollector(root, invocationID, runID);
  let evidence, ctx, signal = "", closing = false, closeOutput = "";
  const publishRows = (results) => {
    if (!request.publish_rows) return;
    for (const result of results) publishRowResult(runRoot, { schema_id: "cartulary.harness_row_result.v6", ...result, runner: "vitest",
      started_at: startedAt, finished_at: new Date().toISOString(), wall_duration_ms: Date.now() - started }, { preserveEquivalent: true });
  };
  const cancel = (value) => { signal ||= value; if (ctx) void ctx.cancelCurrentRun("keyboard-input").catch(() => {}); };
  const onInterrupt = () => cancel("SIGINT"), onTerminate = () => cancel("SIGTERM");
  process.on("SIGINT", onInterrupt); process.on("SIGTERM", onTerminate);
  const stderr = new InvocationStderr({ write(chunk, _encoding, done) {
    if (closing) closeOutput = (closeOutput + chunk.toString()).slice(0, 262144);
    process.stderr.write(chunk); done();
  } });
  try {
    evidence = createVitestEvidence({ root, request, reportFile, detailsFile, runID, invocationID, startedAt });
    const require = createRequire(path.join(root, "apps/web/package.json"));
    const { createVitest, ForksPoolWorker } = await import(pathToFileURL(require.resolve("vitest/node")).href);
    const workers = settledForkPool(ForksPoolWorker);
    ctx = await createVitest("test", {
      root: request.config ? path.dirname(path.resolve(root, request.config)) : path.join(root, "apps/web"),
      config: request.config ? path.resolve(root, request.config) : path.join(root, "apps/web/vite.config.ts"),
      ...(request.project ? { project: [request.project] } : {}), watch: false, run: true, maxWorkers: 1, fileParallelism: false, isolate: true, pool: workers.policy,
      reporters: [collector, ["json", { outputFile: rawReport }]],
    }, {}, { stderr });
    await ctx.standalone();
    const project = ctx.projects.find((candidate) => candidate.name === request.project);
    if (!project || project.config.isolate !== true || (project.config.maxWorkers ?? ctx.config.maxWorkers) !== 1 || ctx.config.maxWorkers !== 1 || ctx.config.fileParallelism !== false || project.config.pool !== workers.policy.name) throw diagnosticFailure(`Vitest project does not satisfy isolated single-worker execution: ${JSON.stringify({ selected: request.project, projects: ctx.projects.map((item) => item.name), isolate: project?.config.isolate, maxWorkers: project?.config.maxWorkers ?? ctx.config.maxWorkers, fileParallelism: ctx.config.fileParallelism, pool: project?.config.pool })}`, "scheduler_accounting_error");
    const specifications = request.files.map((file) => {
      const absolute = path.resolve(root, file.file);
      if (!ctx.getModuleSpecifications(absolute).some((spec) => spec.project === project)) throw diagnosticFailure("Selected Vitest file is excluded from its registered project", "scheduler_accounting_error");
      return project.createSpecification(absolute, { testNamePattern: new RegExp(`^(?:${file.rows.flatMap((row) => row.selectors).map(regexEscape).join("|")})$`) });
    });
    for (let index = 0; index < specifications.length && !signal; index++) {
      collector.latest = undefined;
      const settleFile = workers.beginFile();
      await ctx.runTestSpecifications([specifications[index]]);
      const current = collector.latest;
      if (!current) throw diagnosticFailure("Vitest did not settle a file report");
      if (current.run_status === "interrupted") { signal ||= "SIGINT"; break; }
      const teardownError = await settleFile();
      if (signal) break;
      if (current.completed_files.length !== 1 || current.completed_files[0].owner_path !== request.files[index].file || current.completed_files[0].project !== request.project) throw diagnosticFailure("Vitest settled a different file", "scheduler_accounting_error");
      if (teardownError) current.observations.push({ project: request.project, owner_path: request.files[index].file, title: "(environment teardown)", test_name: "", scope: "suite", status: "failed", duration_ms: 0, timeout_ms: null,
        errors: [{ name: String(teardownError.name ?? "Error"), message: String(teardownError.message ?? ""), stack: String(teardownError.stack ?? "") }] });
      const runObservations = current.observations.filter((item) => item.scope === "run");
      evidence.recordRunFailures(runObservations);
      const settled = evidence.settleFile(index, current, readVitestJSON(rawReport));
      if (runObservations.length) {
        publishRows(adaptVitestInvocation(invocation, evidence.snapshot()));
        break;
      }
      const fileInvocation = vitestInvocation(root, { ...request, files: [request.files[index]] });
      publishRows(adaptVitestInvocation(fileInvocation, settled));
    }
    closing = true;
    await ctx.close();
    if (closeOutput) {
      evidence.recordRunFailures([{ project: "", owner_path: "", title: "(invocation teardown)", test_name: "", scope: "run", status: "failed", duration_ms: 0, timeout_ms: null,
        errors: [{ name: "Error", message: redactString(closeOutput), stack: "" }] }]);
    }
    const terminal = evidence.finish(signal);
    const outcomes = adaptVitestInvocation(invocation, { ...terminal, signal });
    publishRows(outcomes);
    if (signal) return signal === "SIGTERM" ? 143 : 130;
    const failures = outcomes.filter((row) => row.failure_class);
    if (failures.length) {
      const primary = primaryPublicFailure(failures);
      process.stderr.write(`[FAIL] failure_class=${primary.failure_class} failure_reason=${primary.failure_reason} diagnostic=${path.relative(root, detailsFile)}\n`);
      return publicExitCodeForFailures(failures);
    }
    if (terminal.details.run_status !== "passed" || terminal.report.reports.some(({ report }) => report.success !== true)) throw diagnosticFailure("Vitest native status contradicts successful observations", "scheduler_accounting_error");
    return 0;
  } catch (error) {
    const failure = error.failure_reason ? error : diagnosticFailure("Vitest invocation failed to settle", "scheduler_accounting_error");
    evidence?.fail(failure);
    publishRows(adaptVitestInvocation(invocation, { commandFailure: failure }));
    throw error;
  } finally {
    closing = true;
    try { await ctx?.close(); }
    finally {
      process.removeListener("SIGINT", onInterrupt); process.removeListener("SIGTERM", onTerminate);
      stderr.close();
      rmSync(privateRoot, { recursive: true, force: true });
    }
  }
}

try { process.exitCode = await main(); }
catch (error) {
  const failure = error.failure_reason ? error : diagnosticFailure("Vitest invocation could not retain complete diagnostic evidence");
  process.stderr.write(`[FAIL] failure_class=${failure.failure_class} failure_reason=${failure.failure_reason} message=${failure.message} cause=${redactString(error.message ?? "unknown")}\n`);
  process.exitCode = publicExitCodeForFailure(reportCommandFailure(root, error, failure));
}
