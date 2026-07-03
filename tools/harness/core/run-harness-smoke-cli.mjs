#!/usr/bin/env node

import { spawn } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import {
  harnessSummaryGroups,
  harnessSummaryTargets,
  loadSummaryTopologyContext,
  summaryGroupsSpec,
} from "../planning/summary-topology.mjs";
import {
  harnessCheck,
  harnessTierChecks,
  loadTaskSurfaceManifest,
  repoRoot,
} from "../generated-artifacts/task-surface.mjs";
import { verboseOutput as toolVerboseOutput } from "./tool-output.mjs";
import { publicExitCodeForSummary } from "./failure-taxonomy.mjs";

function parseArgs(argv) {
  const options = {
    jobs: process.env.HARNESS_SMOKE_JOBS ?? "1",
    manifest:
      process.env.TASK_SURFACE_MANIFEST ??
      process.env.CARTULARY_TASK_SURFACE_MANIFEST,
    tier: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--tier") {
      options.tier = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--jobs") {
      options.jobs = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--manifest") {
      options.manifest = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    throw new Error(`unknown option ${arg}`);
  }
  if (!options.tier) {
    throw new Error(
      "usage: run-harness-smoke.mjs --tier <tier> [--jobs <count>] [--manifest <path>]",
    );
  }
  return options;
}

function parseJobs(raw) {
  const value = Number.parseInt(raw, 10);
  if (!Number.isFinite(value) || value < 1) {
    throw new Error(
      `HARNESS_SMOKE_JOBS must be a positive integer, got ${JSON.stringify(raw)}`,
    );
  }
  return value;
}

function runCommand(command, args, env = process.env) {
  return new Promise((resolve) => {
    const child = spawn(command, args, {
      cwd: repoRoot,
      env,
      stdio: "inherit",
    });
    child.on("error", (error) => {
      console.error(`${command} failed to start: ${error.message}`);
      resolve(127);
    });
    child.on("close", (status, signal) => {
      if (signal) {
        resolve(1);
        return;
      }
      resolve(status ?? 0);
    });
  });
}

async function emitTestOutput(args) {
  const script =
    process.env.TEST_OUTPUT_SCRIPT ??
    path.join(repoRoot, "tools", "harness", "core", "test-output.mjs");
  if (script.endsWith(".mjs")) {
    return runCommand(process.env.NODE_BIN || process.execPath, [
      script,
      ...args,
    ]);
  }
  return runCommand(script, args);
}

async function summarizeCheck(name, status) {
  return emitTestOutput([
    "target-summary",
    name,
    status === 0 ? "pass" : "fail",
    "--quiet-success",
    "--suppress-machine-output",
  ]);
}

function targetPublicExitCode(target, fallbackStatus) {
  const resultsRoot = process.env.CARTULARY_TEST_RESULTS_DIR ?? path.join(repoRoot, ".cartulary", "test-results");
  const runID = process.env.CARTULARY_TEST_RUN_ID ?? "adhoc";
  const summaryFile = path.join(resultsRoot, runID, target, "tool-run-summary.json");
  if (!existsSync(summaryFile)) {
    return fallbackStatus === 0 ? 0 : 1;
  }
  try {
    return publicExitCodeForSummary(JSON.parse(readFileSync(summaryFile, "utf8")));
  } catch {
    return fallbackStatus === 0 ? 0 : 1;
  }
}

function verboseOutput() {
  return toolVerboseOutput();
}

async function runCheck(check) {
  const runPhase =
    process.env.RUN_PHASE_SCRIPT ??
    path.join(repoRoot, "tools", "harness", "core", "run-phase.sh");
  const command = resolveCheckCommand(check);
  const env = {
    ...process.env,
    CARTULARY_TEST_TARGET: check.name,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  };
  const status = await runCommand(
    runPhase,
    [
      check.name,
      "--",
      "env",
      "-u",
      "CARTULARY_TEST_TARGET",
      "-u",
      "CARTULARY_SUPPRESS_CHILD_SUCCESS",
      "-u",
      "VERBOSE",
      "-u",
      "CI_VERBOSE",
      "CARTULARY_OUTPUT_MODE=summary",
      ...command,
    ],
    env,
  );
  const summaryStatus = await summarizeCheck(check.name, status);
  return summaryStatus === 0
    ? targetPublicExitCode(check.name, status)
    : targetPublicExitCode(check.name, summaryStatus);
}

function resolveCheckCommand(check) {
  const command =
    check.command?.length > 0
      ? check.command
      : [`./${check.backing_scripts[0]}`];
  return command.map((token) => {
    if (token === "$(NODE_BIN)") {
      return process.env.NODE_BIN || process.execPath;
    }
    if (token === "$(MAKE)") {
      return process.env.MAKE || "make";
    }
    return token;
  });
}

async function runChecks(checks, jobs) {
  let next = 0;
  let firstFailure = null;
  let active = 0;
  let resolved = false;

  return new Promise((resolve) => {
    const finishIfDone = () => {
      if (
        resolved ||
        active !== 0 ||
        (next < checks.length && firstFailure === null)
      ) {
        return;
      }
      resolved = true;
      resolve({
        failure: firstFailure,
        skippedAfterFailure: firstFailure
          ? checks.slice(next).map((check) => check.name)
          : [],
      });
    };
    const schedule = () => {
      while (active < jobs && next < checks.length && firstFailure === null) {
        const check = checks[next];
        next += 1;
        active += 1;
        const finishCheck = (status) => {
          active -= 1;
          if (status !== 0 && firstFailure === null) {
            firstFailure = { check: check.name, status };
          }
          schedule();
          finishIfDone();
        };
        void runCheck(check).then(
          (status) => {
            finishCheck(status);
          },
          (error) => {
            console.error(
              `${check.name} failed: ${error instanceof Error ? error.message : String(error)}`,
            );
            finishCheck(1);
          },
        );
      }
      finishIfDone();
    };
    schedule();
  });
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const jobs = parseJobs(options.jobs);
  const { manifest } = loadTaskSurfaceManifest(options.manifest);
  const label = `run-harness-smoke-${options.tier}`;
  const checkNames = harnessTierChecks(manifest, options.tier);
  const checks = checkNames.map((name) => harnessCheck(manifest, name));
  const topologyContext = loadSummaryTopologyContext({
    taskSurfaceManifest: manifest,
  });
  const summaryTargets = harnessSummaryTargets(topologyContext, options.tier);
  const groupsSpec = summaryGroupsSpec(
    harnessSummaryGroups(topologyContext, options.tier),
  );

  if (verboseOutput()) {
    await emitTestOutput([
      "run-start",
      label,
      "--steps",
      "1",
      "--summary-targets",
      String(summaryTargets.length),
      "--helper-units",
      "0",
      "--jobs",
      String(jobs),
    ]);
    await emitTestOutput([
      "step-start",
      label,
      "1",
      "1",
      options.tier,
      "--mode",
      "parallel",
      "--jobs",
      String(jobs),
    ]);
  }

  const { failure, skippedAfterFailure } = await runChecks(checks, jobs);
  const summaryArgs = ["run-summary", label];
  if (failure) {
    summaryArgs.push("fail", "0", "1", failure.check);
  } else {
    summaryArgs.push("pass", "1", "1", "-");
  }
  if (groupsSpec) {
    summaryArgs.push("--summary-groups", groupsSpec);
  }
  if (skippedAfterFailure.length > 0) {
    summaryArgs.push("--skipped-after-failure", skippedAfterFailure.join(","));
  }
  summaryArgs.push("--quiet-success");
  summaryArgs.push("--suppress-machine-output");
  summaryArgs.push(...summaryTargets);
  const summaryStatus = await emitTestOutput(summaryArgs);
  const targetSummaryArgs = [
    "target-summary",
    label,
    failure ? "fail" : "pass",
    "--children",
    summaryTargets.join(","),
    "--quiet-success",
  ];
  if (skippedAfterFailure.length > 0) {
    targetSummaryArgs.push(
      "--skipped-after-failure",
      skippedAfterFailure.join(","),
    );
    if (failure?.check) {
      targetSummaryArgs.push("--failed-dependency", failure.check);
    }
  }
  const targetSummaryStatus = await emitTestOutput(targetSummaryArgs);
  if (failure) {
    process.exit(targetPublicExitCode(failure.check, failure.status));
  }
  process.exit(targetPublicExitCode(label, summaryStatus === 0 ? targetSummaryStatus : summaryStatus));
}

main().catch((error) => {
  console.error(`harness smoke failed: ${error.message}`);
  process.exit(1);
});
