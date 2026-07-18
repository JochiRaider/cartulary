import { spawnSync } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";

import { redactString } from "../contract/index.mjs";

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function regexEscape(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
}

function exactAlternation(values) {
  return `^(?:${values.map(regexEscape).join("|")})$`;
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

function goCommands(rows, workers) {
  const byPackage = new Map();
  for (const row of rows) {
    const values = byPackage.get(row.selector.package) ?? [];
    values.push(...row.selector.tests);
    byPackage.set(row.selector.package, values);
  }
  return [...byPackage.entries()]
    .sort(([left], [right]) => asciiCompare(left, right))
    .map(([packageName, symbols]) => ({
      command: process.env.GO || "go",
      args: [
        "test",
        "-count=1",
        `-p=${workers}`,
        "-run",
        exactAlternation([...new Set(symbols)].sort(asciiCompare)),
        packageName,
      ],
    }));
}

function vitestCommands(root, rows, workers) {
  const byFile = new Map();
  for (const row of rows) {
    const values = byFile.get(row.selector.file) ?? [];
    values.push(...row.selector.titles);
    byFile.set(row.selector.file, values);
  }
  const pnpm = process.env.PNPM || path.join(root, "tmp", "node-runtime", "bin", "pnpm");
  return [...byFile.entries()]
    .sort(([left], [right]) => asciiCompare(left, right))
    .map(([file, titles]) => ({
      command: pnpm,
      args: [
        "--dir",
        "apps/web",
        "exec",
        "vitest",
        "run",
        path.resolve(root, file),
        "-t",
        exactAlternation([...new Set(titles)].sort(asciiCompare)),
        `--maxWorkers=${workers}`,
      ],
    }));
}

function shellCommands(rows) {
  return rows.map((row) => ({
    command: process.env.MAKE || "make",
    args: ["--silent", "--no-print-directory", row.target_name],
  }));
}

function commandsForUnit(root, unit, rows, workers) {
  if (unit.runner === "go") return goCommands(rows, Math.min(workers.vitest, 16));
  if (unit.runner === "vitest") return vitestCommands(root, rows, workers.vitest);
  if (unit.runner === "shell") return shellCommands(rows);
  throw new Error(
    `Playwright owner execution is not active until the catalog-native browser topology workstream: ${unit.work_unit_id}`,
  );
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
    let status = 0;
    let stdout = "";
    let stderr = "";
    let failureReason = null;
    try {
      const commands = commandsForUnit(root, unit, rows, plan.workers);
      for (const rawCommand of commands) {
        const command = commandWithManagedServices(root, unit, rawCommand);
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
        if (result.status !== 0) {
          status = result.status;
          failureReason = "test_assertion_failure";
          break;
        }
      }
    } catch (error) {
      status = 1;
      failureReason = "configuration_error";
      stderr += `${error.message}\n`;
    }
    const durationMs = Number((process.hrtime.bigint() - unitStarted) / 1_000_000n);
    unitResults.push({
      work_unit_id: unit.work_unit_id,
      runner: unit.runner,
      target_name: unit.target_name,
      row_ids: [...unit.row_ids],
      status: status === 0 ? "passed" : "failed",
      exit_code: status,
      duration_ms: durationMs,
      failure_reason: failureReason,
      stdout: redactString(stdout),
      stderr: redactString(stderr),
    });
  }
  return {
    duration_ms: Number((process.hrtime.bigint() - started) / 1_000_000n),
    status: unitResults.every((result) => result.status === "passed") ? "pass" : "fail",
    unit_results: unitResults,
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
