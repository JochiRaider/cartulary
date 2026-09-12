#!/usr/bin/env node
import { spawn } from "node:child_process";
import { randomUUID } from "node:crypto";
import { mkdirSync, rmSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  primaryPublicFailure, publicExitCodeForFailure, publicExitCodeForFailures,
  redactValue, secureWriteFile, validateSchemaSync,
} from "../contract/index.mjs";
import {
  diagnosticFailure, normalizeVitestObservations, readVitestJSON, reconcileVitestReport,
} from "../diagnostics/vitest-failure-details.mjs";
import { reportCommandFailure } from "../runtime/command-failure.mjs";
import { borrowSuiteRuntime } from "../runtime/suite-runtime.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

async function main() {
  const [reportFile, detailsFile, separator, command, ...args] = process.argv.slice(2);
  if (!reportFile || !detailsFile || separator !== "--" || !command) throw diagnosticFailure("Vitest invocation requires report, diagnostics, and a command");
  const runRoot = path.resolve(root, process.env.CARTULARY_TEST_RESULTS_DIR, process.env.CARTULARY_TEST_RUN_ID);
  const runtime = borrowSuiteRuntime({ repoRoot: root, runRoot });
  const invocationID = randomUUID();
  const privateRoot = runtime.privatePath(`vitest-${invocationID}`);
  mkdirSync(privateRoot, { mode: 0o700 });
  const captureFile = path.join(privateRoot, "observations.json");
  const rawReport = path.join(privateRoot, "runner.json");
  const environment = {
    ...process.env,
    CARTULARY_VITEST_CAPTURE_FILE: captureFile,
    CARTULARY_VITEST_INVOCATION_ID: invocationID,
    CARTULARY_VITEST_REPO_ROOT: root,
  };
  let child;
  const forwardSignal = (signal) => child?.kill(signal);
  const onInterrupt = () => forwardSignal("SIGINT");
  const onTerminate = () => forwardSignal("SIGTERM");
  process.on("SIGINT", onInterrupt);
  process.on("SIGTERM", onTerminate);
  try {
    child = spawn(command, [...args,
      "--reporter=json",
      `--reporter=${path.join(root, "tools/harness/diagnostics/vitest-collector.mjs")}`,
      `--outputFile.json=${rawReport}`,
    ], { cwd: root, env: environment, stdio: "inherit" });
    const outcome = await new Promise((resolve, reject) => {
      child.once("error", reject);
      child.once("close", (status, signal) => resolve({ status, signal }));
    });
    if (outcome.signal) return outcome.signal === "SIGINT" ? 130 : 143;
    if ([130, 143].includes(outcome.status)) return outcome.status;
    const report = readVitestJSON(rawReport);
    const details = normalizeVitestObservations(readVitestJSON(captureFile), {
      invocationID, runID: process.env.CARTULARY_TEST_RUN_ID,
      runnerJSON: path.relative(root, reportFile),
      stdoutLog: path.relative(root, path.join(path.dirname(reportFile), "stdout.log")),
      stderrLog: path.relative(root, path.join(path.dirname(reportFile), "stderr.log")),
    });
    reconcileVitestReport(details, report, root);
    const redacted = redactValue(details);
    validateSchemaSync(redacted.schema_id, redacted);
    // Publish only schema-validated, redacted observations and raw runner JSON.
    secureWriteFile(reportFile, `${JSON.stringify(redactValue(report))}\n`, { allowedRoot: runRoot });
    secureWriteFile(detailsFile, `${JSON.stringify(redacted, null, 2)}\n`, { allowedRoot: runRoot });
    if (details.run_status === "interrupted") return 130;
    if (details.failures.length) {
      const primary = primaryPublicFailure(details.failures);
      process.stderr.write(`[FAIL] failure_class=${primary.failure_class} failure_reason=${primary.failure_reason} diagnostic=${path.relative(root, detailsFile)}\n`);
      return publicExitCodeForFailures(details.failures);
    }
    if (outcome.status !== 0 || report.success !== true || details.run_status !== "passed") {
      throw diagnosticFailure("Vitest process status contradicts successful observations", "scheduler_accounting_error");
    }
    return 0;
  } finally {
    process.removeListener("SIGINT", onInterrupt);
    process.removeListener("SIGTERM", onTerminate);
    rmSync(privateRoot, { recursive: true, force: true });
  }
}

try { process.exitCode = await main(); }
catch (error) {
  const failure = error.failure_reason ? error : diagnosticFailure("Vitest invocation could not retain complete diagnostic evidence");
  process.stderr.write(`[FAIL] failure_class=${failure.failure_class} failure_reason=${failure.failure_reason} message=${failure.message}\n`);
  process.exitCode = publicExitCodeForFailure(reportCommandFailure(root, error, failure));
}
