#!/usr/bin/env node
import { beginActivity } from "../runtime/local-activity.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";

import { readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import {
  parseStrictJSON,
  publicExitCodeForFailure,
  redactString,
  redactValue,
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  adaptPlaywrightReport,
  playwrightTestList,
  playwrightGroupExitCode,
  playwrightGroupFailure,
  playwrightReportMaximumBytes,
} from "../execution/runners/playwright.mjs";
import { groupRowsByPerformanceFixture } from "../performance-fixture/index.mjs";
import { publishCommandFailure, reportCommandFailure } from "../runtime/command-failure.mjs";
import { runPrivateCapturedProcess } from "../runtime/private-child-process.mjs";
import { ownedProcess, stopOwnedProcess } from "../runtime/owned-process.mjs";
import { enforcePrivateProcessUmask } from "../runtime/private-process.mjs";
import { loadTestCatalog, resolvePlaywrightSelection } from "../test-catalog/index.mjs";
import { resolveBrowserBatchStage } from "./browser-batch-manifest.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { publishRowResult } from "../execution/runners/row-results.mjs";
import { browserCohort } from "./browser-cohorts.mjs";
import { readPublishedUnit } from "../execution/runners/unit-selection.mjs";
import { attachmentAssignments } from "./browser-session-evidence.mjs";
import { collectFrontendMeasurementObservations } from "./frontend-measurement-evidence.mjs";
import { readRendererAttachment } from "./renderer-attachment.mjs";
import { stageVisualSnapshotCandidate } from "./visual-snapshot-promotion.mjs";
import { removePlaywrightWorkingTraces } from "./playwright-output-cleanup.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function usage() {
  return "usage: browser-catalog-group-cli.mjs --manifest <path> --stage <stage> --partition <partition> --group <group>";
}

function parseArgs(argv) {
  const options = { manifest: "", stage: "", partition: "", group: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--manifest") options.manifest = argv[++index] ?? "";
    else if (arg === "--stage") options.stage = argv[++index] ?? "";
    else if (arg === "--partition") options.partition = argv[++index] ?? "";
    else if (arg === "--group") options.group = argv[++index] ?? "";
    else throw new Error(usage());
  }
  if (!options.manifest || !options.stage || !options.partition || !options.group) throw new Error(usage());
  return options;
}

function runRoot() {
  const results = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!results || !runID) {
    throw new Error("CARTULARY_TEST_RESULTS_DIR and CARTULARY_TEST_RUN_ID are required");
  }
  return path.resolve(root, results, runID);
}

function groupArtifactRoot(target, groupName) {
  const safeName = groupName.replaceAll(/[^a-zA-Z0-9_.-]+/gu, "-");
  return path.join(runRoot(), target, "browser-groups", safeName);
}

function writeRowResults(rowResults, runner, startedAt, finishedAt, wallDurationMs) {
  const rowsRoot = path.join(runRoot(), "rows");
  secureMkdir(rowsRoot);
  for (const row of rowResults) {
    const payload = {
      schema_id: "cartulary.harness_row_result.v6",
      ...row,
      runner,
      started_at: startedAt,
      finished_at: finishedAt,
      wall_duration_ms: wallDurationMs,
    };
    validateSchemaSync(payload.schema_id, payload);
    publishRowResult(runRoot(), payload);
  }
}

function sessionArtifacts() {
  const stackPath = process.env.CARTULARY_WEB_E2E_STACK_JSON_FILE;
  const stackDigest = process.env.CARTULARY_WEB_E2E_STACK_SHA256;
  const diagnosticRef = process.env.CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS_REF;
  const diagnosticDigest =
    process.env.CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS_SHA256;
  if (!stackPath || !stackDigest || !diagnosticRef || !diagnosticDigest) {
    throw new Error("browser group requires validated v4 session artifact evidence");
  }
  return [
    {
      kind: "stack_v7",
      ref: path.relative(runRoot(), stackPath).replaceAll("\\", "/"),
      sha256: stackDigest,
    },
    {
      kind: "startup_diagnostics_v2",
      ref: diagnosticRef,
      sha256: diagnosticDigest,
    },
  ];
}

function validateCurrentSessionAttachment() {
  const stackPath = process.env.CARTULARY_WEB_E2E_STACK_JSON_FILE?.trim() ?? "";
  if (!stackPath) {
    throw new Error(
      "browser group requires CARTULARY_WEB_E2E_STACK_JSON_FILE",
    );
  }
  const assignments = attachmentAssignments(stackPath);
  if (
    !assignments ||
    typeof assignments !== "object" ||
    Array.isArray(assignments) ||
    Object.values(assignments).some((value) => typeof value !== "string")
  ) {
    throw new Error("browser v5 attachment validator returned invalid environment");
  }
  Object.assign(process.env, assignments);
}

function executionTarget(group) {
  const mode = process.env.CARTULARY_BROWSER_MAINTENANCE_MODE ?? "";
  if (mode === "") return group.target;
  if (mode !== "snapshot_update" || group.kind !== "visual") {
    throw new Error(`unsupported browser maintenance mode ${mode || "(empty)"} for ${group.kind}`);
  }
  if (process.env.CARTULARY_TEST_TARGET !== "browser-e2e-visual-update") {
    throw new Error("snapshot_update mode requires CARTULARY_TEST_TARGET=browser-e2e-visual-update");
  }
  return process.env.CARTULARY_TEST_TARGET;
}

function readPlaywrightReport(reportPath) {
  return parseStrictJSON(readLocalFile(reportPath, {
    maximum: playwrightReportMaximumBytes, privateFile: false, allowReplacement: true,
  }).toString("utf8"));
}

function groupRows(catalog, group) {
  const selectedRowIDs = group.selectedRowIDs;
  const rows = selectedRowIDs.map((rowID) => {
    const row = catalog.rowByID.get(rowID);
    if (!row || row.runner !== "playwright") {
      throw new Error(`browser group ${group.name} references non-Playwright catalog row ${rowID}`);
    }
    if (row.runtime_profile_id !== group.runtimeProfileID) {
      throw new Error(`browser group ${group.name} mixes runtime profile ${row.runtime_profile_id}`);
    }
    if ((row.fixture_profile_id ?? "") !== group.fixtureProfileID) {
      throw new Error(
        `browser group ${group.name} fixture profile diverges for ${rowID}`,
      );
    }
    if (!group.specs.includes(row.selector.file)) {
      throw new Error(`browser group ${group.name} omits selector file for ${rowID}`);
    }
    return row;
  });
  const selected = [...selectedRowIDs].sort();
  const resolved = rows.map((row) => row.row_id).sort();
  if (JSON.stringify(selected) !== JSON.stringify(resolved)) {
    throw new Error(`browser group ${group.name} catalog selection is ambiguous`);
  }
  return rows;
}

function commandForGroup(rows, selection, group, artifactRoot) {
  const pnpm = process.env.PNPM || path.join(root, `${workspaceLayout.node_runtime}/bin/pnpm`);
  const selectionFile = path.join(artifactRoot, "playwright-test-list.txt");
  secureWriteFile(selectionFile, playwrightTestList(selection));
  const projectIDs = [...new Set(rows.map((row) => row.selector.project_id))].sort();
  if (projectIDs.length !== 1) {
    throw new Error(`browser group ${group.name} must select exactly one Playwright project`);
  }
  const args = [
    "--dir",
    "apps/web",
    "exec",
    "playwright",
    "test",
    "--config",
    group.config,
    `--reporter=json,${path.join(root, "tools/harness/browser/playwright-cohort-reporter.mjs")}`,
    "--project",
    projectIDs[0],
    "--output",
    path.join(artifactRoot, "playwright-output"),
    ...group.specs,
    "--test-list",
    selectionFile,
  ];
  args.push("--workers=1", "--max-failures=1");
  if (group.kind === "visual" && process.env.CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS === "1") {
    args.push("--update-snapshots=changed");
  }
  return { command: pnpm, args };
}

function visualSnapshotScratch(target) {
  if (target !== "browser-e2e-visual-update") return "";
  const scratch = path.join(runRoot(), target, "snapshot-scratch");
  stageVisualSnapshotCandidate(
    path.join(root, "apps/web/e2e/workbook.visual.spec.ts-snapshots"),
    scratch,
  );
  return scratch;
}

async function main() {
  enforcePrivateProcessUmask();
  const options = parseArgs(process.argv.slice(2));
  const manifest = path.resolve(root, options.manifest);
  const stage = resolveBrowserBatchStage(manifest, options.stage);
  const partition = stage.groups.find((entry) => entry.name === options.partition);
  if (!partition) throw new Error(`browser stage ${options.stage} has no partition ${options.partition}`);
  const catalog = loadTestCatalog(root);
  const unit = readPublishedUnit(runRoot(), process.env.CARTULARY_TEST_RUN_ID, process.env.CARTULARY_WORK_UNIT_ID);
  if (unit.command.args[0] !== "tools/harness/browser/browser-catalog-group-cli.mjs" ||
      JSON.stringify(unit.command.args.slice(1)) !== JSON.stringify(process.argv.slice(2)) ||
      !unit.row_bindings?.length || unit.row_bindings.some((row) => row.runner !== "playwright")) {
    throw new Error("browser selection requires its exact published work unit");
  }
  const group = browserCohort(partition, unit.row_bindings.map(({ row_id }) => catalog.rowByID.get(row_id)), stage.invocationPolicy);
  if (group.name !== options.group) throw new Error("browser cohort identity does not bind selected rows and policy");
  validateCurrentSessionAttachment();
  const attachedSessionID =
    process.env.CARTULARY_BROWSER_SESSION_GROUP?.trim() ?? "";
  const declaredSessionContract =
    process.env.CARTULARY_BROWSER_SESSION_CONTRACT?.trim() ?? attachedSessionID;
  if (
    process.env.CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED !== "1" ||
    !attachedSessionID ||
    declaredSessionContract !== group.browserSessionGroup ||
    process.env.CARTULARY_BROWSER_RUNTIME_PROFILE_ID !==
      group.runtimeProfileID ||
    process.env.CARTULARY_BROWSER_SERVICE_REQUIREMENT !==
      group.serviceRequirement
  ) {
    throw new Error(`browser group ${group.name} session attachment mismatch`);
  }
  const rows = groupRows(catalog, group);
  const snapshotKey = process.env.CARTULARY_FIXTURE_SNAPSHOT_KEY?.trim() ?? "";
  if (
    group.fixtureProfileID &&
    (
      process.env.CARTULARY_FIXTURE_PROFILE_ID !== group.fixtureProfileID ||
      !/^[a-f0-9]{64}$/u.test(snapshotKey)
    )
  ) {
    throw new Error(`browser group ${group.name} snapshot identity is missing or inconsistent`);
  }
  const target = executionTarget(group);
  const artifactRoot = groupArtifactRoot(target, group.name);
  secureMkdir(artifactRoot);
  const reportPath = path.join(artifactRoot, "playwright-report.json");
  const stdoutPath = path.join(artifactRoot, "stdout.log");
  const stderrPath = path.join(artifactRoot, "stderr.log");
  const selection = resolvePlaywrightSelection(root, rows);
  const invocation = commandForGroup(rows, selection, group, artifactRoot);
  const rendererLease = readRendererAttachment(process.env);
  let captureProof, captureStop, signalCleanupError, interruptedSignal;
  const onSignal = (signal) => {
    interruptedSignal ??= signal;
    if (captureProof) captureStop ??= stopOwnedProcess(captureProof).catch((error) => { signalCleanupError = error; });
  };
  const onInterrupt = () => onSignal("SIGINT");
  const onTerminate = () => onSignal("SIGTERM");
  process.once("SIGINT", onInterrupt);
  process.once("SIGTERM", onTerminate);
  const startedAt = new Date().toISOString();
  const started = Date.now();
  let child;
  let primaryError;
  let cleanupError;
  try {
    child = await runPrivateCapturedProcess(invocation.command, invocation.args, {
      producer: "playwright", dependencies: JSON.parse(process.env.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES),
      cwd: root,
      detached: true,
      onSpawn: (process) => { captureProof = ownedProcess(process.pid); if (interruptedSignal) onSignal(interruptedSignal); },
      env: {
        ...process.env,
        ...rendererLease.environment,
        PLAYWRIGHT_JSON_OUTPUT_FILE: reportPath,
        CARTULARY_PLAYWRIGHT_WORKER_COUNT: "1",
        CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET: "0",
        CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
        CARTULARY_VISUAL_SNAPSHOT_ROOT: visualSnapshotScratch(target),
        CARTULARY_FRONTEND_ACCESSIBILITY_CONTRAST_DIR:
          group.kind === "a11y" ? path.join(artifactRoot, "contrast-checks") : "",
      },
      repoRoot: root,
      runRoot: runRoot(),
    });
    secureWriteFile(stdoutPath, redactString(child.stdout ?? ""));
    secureWriteFile(stderrPath, redactString(child.stderr ?? ""));
  } catch (error) {
    primaryError = error;
    throw error;
  } finally {
    await captureStop;
    cleanupError ??= signalCleanupError;
    try {
      if (captureProof) await stopOwnedProcess(captureProof);
      removePlaywrightWorkingTraces(path.join(artifactRoot, "playwright-output"));
    } catch (error) { cleanupError ??= error; }
    for (const release of [() => child?.cleanup()]) {
      try { release(); }
      catch (error) { cleanupError ??= error; }
    }
    if (cleanupError) process.stderr.write("private browser resources cleanup failed (cleanup_error)\n");
    process.removeListener("SIGINT", onInterrupt);
    process.removeListener("SIGTERM", onTerminate);
    if (cleanupError && !primaryError && !child) throw cleanupError;
  }
  if (interruptedSignal) child.signal ??= interruptedSignal;
  let report = null;
  const parsing = beginActivity({ repoRoot: root, environment: process.env, launch: child.launch, activity: "report_parse" });
  try {
    report = readPlaywrightReport(reportPath);
    secureWriteFile(reportPath, `${JSON.stringify(redactValue(report), null, 2)}\n`);
  } catch {
    try { report = readPlaywrightReport(path.join(artifactRoot, "playwright-checkpoint.json")); }
    catch { report = null; }
  }
  let rowResults;
  try { rowResults = adaptPlaywrightReport(
    rows,
    report,
    child.status,
    child.signal,
    selection,
  ); parsing.finish(!report || rowResults.some((row) => row.failure_class === "artifact" || row.failure_class === "harness") ? "failed" : "passed"); }
  catch (error) { parsing.finish("failed"); throw error; }
  let measurementEvidenceError = null;
  const fixtureGroups = groupRowsByPerformanceFixture(root, rows);
  if (fixtureGroups.length > 0) {
    try {
      for (const fixtureGroup of fixtureGroups) {
        collectFrontendMeasurementObservations({
          expectedPredicateIDs: fixtureGroup.predicate_ids,
          observationSchemaID: fixtureGroup.profile.artifact_policy.observation_schema_id,
          reportPaths: [reportPath],
          runRoot: runRoot(),
        });
      }
    } catch (error) {
      measurementEvidenceError = error instanceof Error ? error : new Error(String(error));
      secureWriteFile(
        stderrPath,
        `${readFileSync(stderrPath, "utf8")}${redactString(`frontend measurement evidence invalid: ${measurementEvidenceError.message}\n`)}`,
      );
    }
  }
  const primaryExitCode = measurementEvidenceError === null
    ? playwrightGroupExitCode(rowResults, child, report)
    : 11;
  const exitCode = primaryExitCode || (cleanupError ? 12 : 0);
  const finishedAt = new Date().toISOString();
  const wallDurationMs = Date.now() - started;
  writeRowResults(rowResults, "playwright", startedAt, finishedAt, wallDurationMs);
  const result = {
    schema_id: "cartulary.browser_group_result.v11",
    renderer_borrow: { unit_id: rendererLease.unit_id, borrow_id: rendererLease.borrow_id, allocation_id: rendererLease.allocation_id, generation: rendererLease.generation, attestation_ref: rendererLease.attestation_ref, trust_ref: rendererLease.trust_ref },
    target_id: target,
    stage_id: stage.name,
    group_id: group.name,
    // Report the concrete run-scoped lease identity, not the authored logical
    // session contract. Non-stateful groups may intentionally declare the same
    // logical contract while using independent stacks; target projections must
    // never collapse those distinct resources into one session.
    browser_session_id: attachedSessionID,
    runtime_profile_id: group.runtimeProfileID,
    service_requirement: group.serviceRequirement,
    fixture_capabilities: [...new Set(rows.map((row) => row.fixture_capability))].sort(),
    service_dependencies: [...new Set(rows.flatMap((row) => row.service_dependencies))].sort(),
    resource_profile_ids: [...new Set(rows.map((row) => row.resource_profile_id))].sort(),
    ...(group.fixtureProfileID
      ? {
          fixture_profile_id: group.fixtureProfileID,
          snapshot_key: snapshotKey,
        }
      : {}),
    selected_rows: rows.map((row) => row.row_id).sort(),
    started_at: startedAt,
    finished_at: finishedAt,
    duration_ms: wallDurationMs,
    status: exitCode === 0 ? "pass" : "fail",
    exit_code: exitCode,
    row_results: rowResults,
    session_artifacts: sessionArtifacts(),
    artifacts: {
      playwright_report: path.relative(runRoot(), reportPath).replaceAll("\\", "/"),
      stdout: path.relative(runRoot(), stdoutPath).replaceAll("\\", "/"),
      stderr: path.relative(runRoot(), stderrPath).replaceAll("\\", "/"),
    },
  };
  validateSchemaSync(result.schema_id, result);
  secureWriteFile(path.join(artifactRoot, "browser-group-result.json"), `${JSON.stringify(result, null, 2)}\n`);
  if (exitCode === 12 && cleanupError) process.stderr.write("failure_class=harness failure_reason=cleanup_error\n");
  if (exitCode !== 0) {
    const failure = measurementEvidenceError
      ? { failure_class: "artifact", failure_reason: "artifact_error" }
      : playwrightGroupFailure(rowResults, child, report) ?? { failure_class: "harness", failure_reason: "cleanup_error" };
    publishCommandFailure(root, failure);
    process.stderr.write(`failure_class=${failure.failure_class} failure_reason=${failure.failure_reason}\n`);
    process.stderr.write(
      `[FAIL] target=${group.target} group=${group.name} exit_code=${exitCode} failed_rows=${rowResults.filter((row) => row.terminal_state !== "passed").length}\n`,
    );
  }
  return exitCode;
}

try {
  process.exitCode = await main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  if (error.message === usage()) process.exitCode = 2;
  else process.exitCode = publicExitCodeForFailure(reportCommandFailure(root, error, { failure_class: "artifact", failure_reason: "artifact_error" }));
}
