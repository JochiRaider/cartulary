#!/usr/bin/env node

import { spawn } from "node:child_process";
import path from "node:path";
import {
  harnessCheck,
  harnessTierChecks,
  loadTaskSurfaceManifest,
  repoRoot,
  summaryProfileArgs,
} from "./lib/task-surface.mjs";

function parseArgs(argv) {
  const options = {
    jobs: process.env.HARNESS_SMOKE_JOBS ?? "1",
    manifest: process.env.TASK_SURFACE_MANIFEST ?? process.env.CARTULARY_TASK_SURFACE_MANIFEST,
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
    throw new Error("usage: run-harness-smoke.mjs --tier <tier> [--jobs <count>] [--manifest <path>]");
  }
  return options;
}

function parseJobs(raw) {
  const value = Number.parseInt(raw, 10);
  if (!Number.isFinite(value) || value < 1) {
    throw new Error(`HARNESS_SMOKE_JOBS must be a positive integer, got ${JSON.stringify(raw)}`);
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
  const script = process.env.TEST_OUTPUT_SCRIPT ?? path.join(repoRoot, "scripts", "lib", "test-output.sh");
  return runCommand(script, args);
}

async function summarizeCheck(name, status) {
  return emitTestOutput(["target-summary", name, status === 0 ? "pass" : "fail", "--quiet-success"]);
}

function verboseOutput() {
  return (
    process.env.VERBOSE === "1" ||
    process.env.CI_VERBOSE === "1" ||
    (process.env.CARTULARY_OUTPUT_MODE ?? "quiet") !== "quiet"
  );
}

async function runCheck(check) {
  const runPhase = path.join(repoRoot, "scripts", "lib", "run-phase.sh");
  const script = check.backing_scripts[0];
  const env = {
    ...process.env,
    CARTULARY_TEST_TARGET: check.name,
  };
  const status = await runCommand(runPhase, [
    check.name,
    "--",
    "env",
    "-u",
    "CARTULARY_TEST_TARGET",
    `./${script}`,
  ], env);
  const summaryStatus = await summarizeCheck(check.name, status);
  return status === 0 ? summaryStatus : status;
}

async function runChecks(checks, jobs) {
  let next = 0;
  let firstFailure = null;
  let active = 0;
  let resolved = false;

  return new Promise((resolve) => {
    const finishIfDone = () => {
      if (resolved || active !== 0 || (next < checks.length && firstFailure === null)) {
        return;
      }
      resolved = true;
      resolve({
        failure: firstFailure,
        skippedAfterFailure: firstFailure ? checks.slice(next).map((check) => check.name) : [],
      });
    };
    const schedule = () => {
      while (active < jobs && next < checks.length && firstFailure === null) {
        const check = checks[next];
        next += 1;
        active += 1;
        runCheck(check).then((status) => {
          active -= 1;
          if (status !== 0 && firstFailure === null) {
            firstFailure = { check: check.name, status };
          }
          schedule();
          finishIfDone();
        });
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
  const profile = summaryProfileArgs(manifest, label);

  if (verboseOutput()) {
    await emitTestOutput(["run-start", label, "--steps", "1", "--targets", String(checks.length), "--jobs", String(jobs)]);
    await emitTestOutput(["step-start", label, "1", "1", options.tier, "--mode", "parallel", "--jobs", String(jobs)]);
  }

  const { failure, skippedAfterFailure } = await runChecks(checks, jobs);
  const summaryArgs = ["run-summary", label];
  if (failure) {
    summaryArgs.push("fail", "0", "1", failure.check);
  } else {
    summaryArgs.push("pass", "1", "1", "-");
  }
  if (profile.groupsSpec) {
    summaryArgs.push("--summary-groups", profile.groupsSpec);
  }
  if (skippedAfterFailure.length > 0) {
    summaryArgs.push("--skipped-after-failure", skippedAfterFailure.join(","));
  }
  summaryArgs.push("--quiet-success");
  summaryArgs.push(...profile.targets);
  const summaryStatus = await emitTestOutput(summaryArgs);
  if (failure) {
    process.exit(failure.status);
  }
  process.exit(summaryStatus);
}

main().catch((error) => {
  console.error(`harness smoke failed: ${error.message}`);
  process.exit(1);
});
