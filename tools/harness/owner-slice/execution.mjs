import { spawnSync } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";

import { redactString } from "../contract/index.mjs";
import { adaptGoInvocation, buildGoInvocations } from "../execution/runners/go.mjs";
import { adaptShellInvocation, buildShellInvocations } from "../execution/runners/shell.mjs";
import { adaptVitestInvocation, buildVitestInvocations } from "../execution/runners/vitest.mjs";

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

function invocationsForUnit(root, unit, rows, workers) {
  const pnpm = process.env.PNPM || path.join(root, "tmp", "node-runtime", "bin", "pnpm");
  if (unit.runner === "go") return buildGoInvocations(rows, workers.vitest);
  if (unit.runner === "vitest") return buildVitestInvocations(root, rows, workers.vitest, pnpm);
  if (unit.runner === "shell") return buildShellInvocations(rows);
  throw new Error(
    `Playwright owner execution is not active until the catalog-native browser topology workstream: ${unit.work_unit_id}`,
  );
}

function adaptInvocation(runner, invocation, result) {
  if (runner === "go") return adaptGoInvocation(invocation, result);
  if (runner === "vitest") return adaptVitestInvocation(invocation, result);
  if (runner === "shell") return adaptShellInvocation(invocation, result);
  throw new Error(`unsupported owner-slice runner ${runner}`);
}

function commandWithManagedServices(root, unit, command) {
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
      const invocations = invocationsForUnit(root, unit, rows, plan.workers);
      for (const rawInvocation of invocations) {
        const invocationStarted = process.hrtime.bigint();
        const command = commandWithManagedServices(root, unit, rawInvocation);
        const result = (options.runCommand ?? run)(command.command, command.args, {
          cwd: root,
          env: {
            ...process.env,
            CARTULARY_TEST_OWNER: plan.owner_id,
            CARTULARY_TEST_CATALOG_ROW_IDS: unit.row_ids.join(","),
            CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
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
