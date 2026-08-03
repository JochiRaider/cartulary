#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import {
  redactString,
  redactValue,
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import { adaptPlaywrightReport } from "../execution/runners/playwright.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";
import { resolveBrowserBatchStage } from "./browser-batch-manifest.mjs";
import { selectedBrowserGroupRowIDs } from "./browser-group-selection.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function usage() {
  return "usage: browser-catalog-group-cli.mjs --manifest <path> --stage <stage> --group <group>";
}

function parseArgs(argv) {
  const options = { manifest: "", stage: "", group: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--manifest") options.manifest = argv[++index] ?? "";
    else if (arg === "--stage") options.stage = argv[++index] ?? "";
    else if (arg === "--group") options.group = argv[++index] ?? "";
    else throw new Error(usage());
  }
  if (!options.manifest || !options.stage || !options.group) throw new Error(usage());
  return options;
}

function regexEscape(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
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
      schema_id: "cartulary.harness_row_result.v1",
      ...row,
      runner,
      started_at: startedAt,
      finished_at: finishedAt,
      wall_duration_ms: wallDurationMs,
    };
    validateSchemaSync(payload.schema_id, payload);
    secureWriteFile(path.join(rowsRoot, `${row.row_id}.json`), `${JSON.stringify(payload, null, 2)}\n`);
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
      kind: "stack_v4",
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
  const validation = spawnSync(
    process.env.NODE_BIN || process.execPath,
    [
      path.join(
        root,
        "tools/harness/browser/browser-session-evidence.mjs",
      ),
      "attach-json",
      stackPath,
    ],
    {
      cwd: root,
      env: process.env,
      encoding: "utf8",
      maxBuffer: 8 * 1024 * 1024,
    },
  );
  if (validation.status !== 0) {
    throw new Error(
      `browser v4 attachment validation failed: ${(validation.stderr || validation.stdout || "unknown validation failure").trim()}`,
    );
  }
  const assignments = JSON.parse(validation.stdout);
  if (
    !assignments ||
    typeof assignments !== "object" ||
    Array.isArray(assignments) ||
    Object.values(assignments).some((value) => typeof value !== "string")
  ) {
    throw new Error("browser v4 attachment validator returned invalid environment");
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

function groupRows(catalog, group) {
  const selectedRowIDs = selectedBrowserGroupRowIDs(group);
  const rows = selectedRowIDs.map((rowID) => {
    const row = catalog.rowByID.get(rowID);
    if (!row || row.runner !== "playwright") {
      throw new Error(`browser group ${group.name} references non-Playwright catalog row ${rowID}`);
    }
    if (row.runtime_profile_id !== group.runtimeProfileID) {
      throw new Error(`browser group ${group.name} mixes runtime profile ${row.runtime_profile_id}`);
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

function commandForGroup(rows, group, artifactRoot) {
  const pnpm = process.env.PNPM || path.join(root, "tmp/node-runtime/bin/pnpm");
  const titles = [...new Set(rows.flatMap((row) => row.selector.titles))].sort();
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
    "playwright.config.ts",
    "--reporter=json",
    "--project",
    projectIDs[0],
    "--output",
    path.join(artifactRoot, "playwright-output"),
    ...group.specs,
    "-g",
    `(?:${titles.map(regexEscape).join("|")})`,
  ];
  if (group.workers !== "default") args.push(`--workers=${group.workers}`);
  if (group.kind === "visual" && process.env.CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS === "1") {
    args.push("--update-snapshots=all");
  }
  return { command: pnpm, args };
}

function exitCodeForRows(rowResults, child) {
  if (child.signal === "SIGINT" || child.signal === "SIGTERM") return 130;
  if (rowResults.some((row) => row.exit_code === 13)) return 13;
  if (rowResults.some((row) => row.terminal_state === "infrastructure_failed")) return 11;
  if (rowResults.some((row) => row.terminal_state === "failed")) return 10;
  return child.status === 0 ? 0 : 11;
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const manifest = path.resolve(root, options.manifest);
  const stage = resolveBrowserBatchStage(manifest, options.stage);
  const group = stage.groups.find((entry) => entry.name === options.group);
  if (!group) throw new Error(`browser stage ${options.stage} has no group ${options.group}`);
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
  const catalog = loadTestCatalog(root);
  const rows = groupRows(catalog, group);
  const target = executionTarget(group);
  const artifactRoot = groupArtifactRoot(target, group.name);
  secureMkdir(artifactRoot);
  const reportPath = path.join(artifactRoot, "playwright-report.json");
  const stdoutPath = path.join(artifactRoot, "stdout.log");
  const stderrPath = path.join(artifactRoot, "stderr.log");
  const invocation = commandForGroup(rows, group, artifactRoot);
  const startedAt = new Date().toISOString();
  const started = Date.now();
  const child = spawnSync(invocation.command, invocation.args, {
    cwd: root,
    env: {
      ...process.env,
      PLAYWRIGHT_JSON_OUTPUT_FILE: reportPath,
      PLAYWRIGHT_WORKERS: group.workers === "default" ? (process.env.PLAYWRIGHT_WORKERS || "2") : group.workers,
      CARTULARY_PLAYWRIGHT_WORKER_COUNT:
        process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT ||
        (group.workers === "default" ? (process.env.PLAYWRIGHT_WORKERS || "2") : group.workers),
      CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET:
        process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET || "0",
      CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
      CARTULARY_FRONTEND_ACCESSIBILITY_CONTRAST_DIR:
        group.kind === "a11y" ? path.join(artifactRoot, "contrast-checks") : "",
    },
    encoding: "utf8",
    maxBuffer: 128 * 1024 * 1024,
  });
  secureWriteFile(stdoutPath, redactString(child.stdout ?? ""));
  secureWriteFile(stderrPath, redactString(child.stderr ?? ""));
  let report = null;
  try {
    report = JSON.parse(readFileSync(reportPath, "utf8"));
    secureWriteFile(reportPath, `${JSON.stringify(redactValue(report), null, 2)}\n`);
  } catch {
    report = null;
  }
  const rowResults = adaptPlaywrightReport(rows, report, child.status ?? 11);
  const exitCode = exitCodeForRows(rowResults, child);
  const finishedAt = new Date().toISOString();
  const wallDurationMs = Date.now() - started;
  writeRowResults(rowResults, "playwright", startedAt, finishedAt, wallDurationMs);
  const result = {
    schema_id: "cartulary.browser_group_result.v3",
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
    resource_profile_ids: [...new Set(rows.map((row) => row.resource_profile_id))].sort(),
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
  if (exitCode !== 0) {
    process.stderr.write(
      `[FAIL] target=${group.target} group=${group.name} exit_code=${exitCode} failed_rows=${rowResults.filter((row) => row.terminal_state !== "passed").length}\n`,
    );
  }
  return exitCode;
}

try {
  process.exitCode = main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = error.message === usage() ? 2 : 11;
}
