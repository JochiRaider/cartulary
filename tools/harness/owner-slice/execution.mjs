import { spawnSync } from "node:child_process";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

import { redactString } from "../contract/index.mjs";
import { adaptGoInvocation, buildGoInvocations } from "../execution/runners/go.mjs";
import { adaptShellInvocation, buildShellInvocations } from "../execution/runners/shell.mjs";
import { adaptVitestInvocation, buildVitestInvocations } from "../execution/runners/vitest.mjs";
import { adaptPlaywrightReport } from "../execution/runners/playwright.mjs";

function regexEscape(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
}

function buildPlaywrightInvocation(root, unit, rows, workers, artifactRoot) {
  const pnpm = process.env.PNPM || path.join(root, "tmp", "node-runtime", "bin", "pnpm");
  const files = [...new Set(rows.map((row) => row.selector.file))].sort();
  const titles = [...new Set(rows.flatMap((row) => row.selector.titles))].sort();
  const projects = [...new Set(rows.map((row) => row.selector.project_id))].sort();
  if (projects.length !== 1) throw new Error(`${unit.work_unit_id} must resolve one Playwright project`);
  mkdirSync(artifactRoot, { recursive: true });
  const reportPath = path.join(artifactRoot, "playwright-report.json");
  return [{
    command: pnpm,
    args: [
      "--dir", "apps/web", "exec", "playwright", "test",
      "--config", "playwright.config.ts",
      "--reporter=json",
      "--project", projects[0],
      "--workers", String(workers.playwright),
      "--output", path.join(artifactRoot, "playwright-output"),
      ...files,
      "-g", `(?:${titles.map(regexEscape).join("|")})`,
    ],
    reportPath,
    rows,
  }];
}

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd,
    env: options.env,
    encoding: "utf8",
    maxBuffer: 64 * 1024 * 1024,
  });
  if (result.error) throw result.error;
  return {
    status: result.status ?? 1,
    stdout: result.stdout ?? "",
    stderr: result.stderr ?? "",
  };
}

function invocationsForUnit(root, unit, rows, workers, artifactRoot) {
  const pnpm = process.env.PNPM || path.join(root, "tmp", "node-runtime", "bin", "pnpm");
  if (unit.runner === "go") return buildGoInvocations(rows, workers.vitest);
  if (unit.runner === "vitest") return buildVitestInvocations(root, rows, workers.vitest, pnpm);
  if (unit.runner === "shell") return buildShellInvocations(rows);
  if (unit.runner === "playwright") {
    return buildPlaywrightInvocation(root, unit, rows, workers, artifactRoot);
  }
  throw new Error(`unsupported owner-slice runner ${unit.runner}`);
}

function adaptInvocation(runner, invocation, result) {
  if (runner === "go") return adaptGoInvocation(invocation, result);
  if (runner === "vitest") return adaptVitestInvocation(invocation, result);
  if (runner === "shell") return adaptShellInvocation(invocation, result);
  if (runner === "playwright") {
    let report = null;
    try {
      report = JSON.parse(readFileSync(invocation.reportPath, "utf8"));
    } catch {
      report = null;
    }
    return adaptPlaywrightReport(invocation.rows, report, result.status);
  }
  throw new Error(`unsupported owner-slice runner ${runner}`);
}

function commandWithManagedServices(root, unit, command) {
  if (unit.runner === "playwright") {
    return {
      command: path.join(root, "tools", "harness", "browser", "start-web-e2e.sh"),
      args: ["--", command.command, ...command.args],
    };
  }
  if (unit.managed_service_ids.length === 0) return command;
  const testServices = process.env.TEST_SERVICES_BIN || path.join(root, "tmp", "toolbin", "cartulary-test-services");
  return {
    command: testServices,
    args: ["run", "--", command.command, ...command.args],
  };
}

export function executeOwnerSlicePlan(root, plan, options = {}) {
  const rowByID = new Map(plan.rows.map((row) => [row.row_id, row]));
  const unitResults = [];
  const started = process.hrtime.bigint();
  for (const unit of plan.work_units) {
    const unitStarted = process.hrtime.bigint();
    const rows = unit.row_ids.map((rowID) => rowByID.get(rowID));
    let stdout = "";
    let stderr = "";
    const rowResults = [];
    try {
      const artifactRoot = path.join(
        options.artifactRoot ?? path.join(root, ".cartulary", "test-results", plan.run_id, plan.target, "work-unit-artifacts"),
        unit.work_unit_id,
      );
      const invocations = invocationsForUnit(root, unit, rows, plan.workers, artifactRoot);
      for (const rawInvocation of invocations) {
        const invocationStarted = process.hrtime.bigint();
        const command = commandWithManagedServices(root, unit, rawInvocation);
        const result = (options.runCommand ?? run)(command.command, command.args, {
          cwd: root,
          env: {
            ...process.env,
            CARTULARY_TEST_OWNER: plan.owner_id,
            CARTULARY_TEST_CATALOG_ROW_IDS: unit.row_ids.join(","),
            // Exact adapters consume the child runner stream. The enclosing
            // scheduler already retains/redacts the unit result and logs.
            CARTULARY_SUPPRESS_CHILD_SUCCESS: "0",
            CARTULARY_TEST_TARGET: plan.target,
            CARTULARY_BROWSER_SESSION_GROUP: `owner-${unit.work_unit_id}`,
            CARTULARY_BROWSER_RUNTIME_PROFILE_ID: unit.runtime_profile_id,
            CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER: unit.runner === "playwright" ? "1" : "",
            PLAYWRIGHT_JSON_OUTPUT_FILE: rawInvocation.reportPath ?? "",
            PLAYWRIGHT_WORKERS: String(plan.workers.playwright),
            CARTULARY_PLAYWRIGHT_WORKER_COUNT: String(plan.workers.playwright),
            CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET: "0",
          },
        });
        stdout += result.stdout;
        stderr += result.stderr;
        const invocationDuration = Number((process.hrtime.bigint() - invocationStarted) / 1_000_000n);
        for (const adapted of adaptInvocation(unit.runner, rawInvocation, result)) {
          rowResults.push({
            ...adapted,
            duration_ms: adapted.duration_ms || invocationDuration,
            attempt: 1,
          });
        }
      }
    } catch (error) {
      stderr += `${error.message}\n`;
      const observed = new Set(rowResults.map((entry) => entry.row_id));
      for (const rowID of unit.row_ids.filter((entry) => !observed.has(entry))) {
        rowResults.push({
          row_id: rowID,
          terminal_state: "infrastructure_failed",
          duration_ms: 0,
          exit_code: 1,
          failure_reason: "runner_adapter_error",
          attempt: 1,
        });
      }
    }
    const durationMs = Number((process.hrtime.bigint() - unitStarted) / 1_000_000n);
    rowResults.sort((left, right) => left.row_id < right.row_id ? -1 : left.row_id > right.row_id ? 1 : 0);
    const states = rowResults.map((entry) => entry.terminal_state);
    const terminalState = states.includes("failed")
      ? "failed"
      : states.includes("infrastructure_failed")
        ? "infrastructure_failed"
        : "passed";
    const exitCode = rowResults.find((entry) => entry.exit_code !== 0)?.exit_code ?? 0;
    unitResults.push({
      work_unit_id: unit.work_unit_id,
      runner: unit.runner,
      target_name: unit.target_name,
      row_ids: [...unit.row_ids],
      status: terminalState,
      exit_code: exitCode,
      duration_ms: durationMs,
      failure_reason: rowResults.find((entry) => entry.failure_reason)?.failure_reason ?? null,
      row_results: rowResults,
      stdout: redactString(stdout),
      stderr: redactString(stderr),
    });
  }
  return {
    duration_ms: Number((process.hrtime.bigint() - started) / 1_000_000n),
    status: unitResults.every((result) => result.status === "passed") ? "pass" : "fail",
    unit_results: unitResults,
    row_results: unitResults.flatMap((result) => result.row_results),
  };
}

export function executeOwnerSliceUnit(root, plan, workUnitID, options = {}) {
  const unit = plan.work_units.find((entry) => entry.work_unit_id === workUnitID);
  if (!unit) throw new Error(`owner slice plan has no work unit ${workUnitID}`);
  return executeOwnerSlicePlan(root, { ...plan, work_units: [unit] }, options);
}

export function retainOwnerSliceUnitLogs(targetDir, execution) {
  const logDir = path.join(targetDir, "work-units");
  mkdirSync(logDir, { recursive: true });
  const retained = [];
  for (const result of execution.unit_results) {
    const base = path.join(logDir, result.work_unit_id);
    const stdoutPath = `${base}.stdout.log`;
    const stderrPath = `${base}.stderr.log`;
    writeFileSync(stdoutPath, result.stdout, "utf8");
    writeFileSync(stderrPath, result.stderr, "utf8");
    retained.push({
      work_unit_id: result.work_unit_id,
      stdout_path: path.relative(path.dirname(targetDir), stdoutPath).replaceAll("\\", "/"),
      stderr_path: path.relative(path.dirname(targetDir), stderrPath).replaceAll("\\", "/"),
    });
  }
  return retained;
}
