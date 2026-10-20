#!/usr/bin/env node
import { randomUUID } from "node:crypto";
import { createRequire } from "node:module";
import { mkdirSync, rmSync } from "node:fs";
import { Writable } from "node:stream";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { primaryPublicFailure, publicExitCodeForFailure, publicExitCodeForFailures, redactString, redactValue, validateSchemaSync } from "../contract/index.mjs";
import { diagnosticFailure, normalizeVitestObservations, readVitestJSON, reconcileVitestReport } from "../diagnostics/vitest-failure-details.mjs";
import VitestCollector from "../diagnostics/vitest-collector.mjs";
import { adaptVitestInvocation, validateVitestRequest, vitestInvocation } from "./runners/vitest.mjs";
import { publishRowResult } from "./runners/row-results.mjs";
import { settledForkPool } from "./runners/vitest-worker-lifecycle.mjs";
import { atomicLocalFile } from "../runtime/secure-local-files.mjs";
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
  const reports = { schema_id: "cartulary.vitest_runner_reports.v1", invocation_id: invocationID, run_id: runID, reports: [] };
  const capture = { schema_id: "cartulary.vitest_failure_details.v3", invocation_id: invocationID, run_id: runID,
    run_status: "passed", generated_at: startedAt, runner_json: "", stdout_log: "", stderr_log: "", observations: [], failures: [], completed_files: [], final_complete: false };
  let ctx, signal = "", closing = false, closeOutput = "", checkpoint = 0, priorReport, details;
  const publishRows = (results) => {
    if (!request.publish_rows) return;
    for (const result of results) publishRowResult(runRoot, { schema_id: "cartulary.harness_row_result.v6", ...result, runner: "vitest",
      started_at: startedAt, finished_at: new Date().toISOString(), wall_duration_ms: Date.now() - started });
  };
  const publish = (finalComplete) => {
    capture.final_complete = finalComplete;
    capture.generated_at = new Date().toISOString();
    const destination = finalComplete ? reportFile : path.join(path.dirname(reportFile), `runner-checkpoint-${++checkpoint}.json`);
    details = normalizeVitestObservations(capture, { invocationID, runID, runnerJSON: path.relative(root, destination),
      stdoutLog: path.relative(root, path.join(path.dirname(reportFile), "stdout.log")), stderrLog: path.relative(root, path.join(path.dirname(reportFile), "stderr.log")) });
    reconcileVitestReport(details, reports, root);
    const redacted = redactValue(details);
    validateSchemaSync(redacted.schema_id, redacted);
    atomicLocalFile(destination, `${JSON.stringify(redactValue(reports))}\n`);
    atomicLocalFile(detailsFile, `${JSON.stringify(redacted, null, 2)}\n`, { replace: true, maximumReplacementBytes: 16 * 1024 * 1024 });
    if (priorReport) rmSync(priorReport);
    priorReport = destination;
    const results = adaptVitestInvocation(invocation, { details, report: reports, signal });
    const completed = new Set(capture.completed_files.map((file) => file.owner_path));
    publishRows(finalComplete ? results : results.filter((result) => completed.has(invocation.rows.find((row) => row.row_id === result.row_id).file)));
  };
  const cancel = (value) => { signal ||= value; if (ctx) void ctx.cancelCurrentRun("keyboard-input").catch(() => {}); };
  const onInterrupt = () => cancel("SIGINT"), onTerminate = () => cancel("SIGTERM");
  process.on("SIGINT", onInterrupt); process.on("SIGTERM", onTerminate);
  const stderr = new InvocationStderr({ write(chunk, _encoding, done) {
    if (closing) closeOutput = (closeOutput + chunk.toString()).slice(0, 262144);
    process.stderr.write(chunk); done();
  } });
  try {
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
      capture.observations.push(...current.observations);
      capture.completed_files.push(...current.completed_files);
      reports.reports.push({ project: request.project, file: request.files[index].file, report: readVitestJSON(rawReport) });
      if (current.run_status === "failed" || current.observations.some((item) => item.status === "failed")) capture.run_status = "failed";
      publish(false);
      if (current.observations.some((item) => item.scope === "run")) break;
    }
    closing = true;
    await ctx.close();
    if (closeOutput) {
      capture.run_status = "failed";
      capture.observations.push({ project: "", owner_path: "", title: "(invocation teardown)", test_name: "", scope: "run", status: "failed", duration_ms: 0, timeout_ms: null,
        errors: [{ name: "Error", message: redactString(closeOutput), stack: "" }] });
    }
    if (signal) capture.run_status = "interrupted";
    publish(!signal);
    if (signal) return signal === "SIGTERM" ? 143 : 130;
    const outcomes = adaptVitestInvocation(invocation, { details, report: reports });
    const failures = outcomes.filter((row) => row.failure_class);
    if (failures.length) {
      const primary = primaryPublicFailure(failures);
      process.stderr.write(`[FAIL] failure_class=${primary.failure_class} failure_reason=${primary.failure_reason} diagnostic=${path.relative(root, detailsFile)}\n`);
      return publicExitCodeForFailures(failures);
    }
    if (details.run_status !== "passed" || reports.reports.some(({ report }) => report.success !== true)) throw diagnosticFailure("Vitest native status contradicts successful observations", "scheduler_accounting_error");
    return 0;
  } catch (error) {
    const failure = error.failure_reason ? error : diagnosticFailure("Vitest invocation failed to settle", "scheduler_accounting_error");
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
