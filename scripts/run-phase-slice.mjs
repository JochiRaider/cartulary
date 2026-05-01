#!/usr/bin/env node
import { accessSync, constants, existsSync, readFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { phaseSlice } from "./lib/task-guidance.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const validModes = new Set(["phase", "service-backed"]);

function usage() {
  process.stderr.write("usage: run-phase-slice.mjs --phase <phaseN> --mode <phase|service-backed>\n");
  process.exit(2);
}

function parseArgs(argv) {
  const options = { phase: "", mode: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--phase") {
      options.phase = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--mode") {
      options.mode = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.phase || !validModes.has(options.mode)) {
    usage();
  }
  return options;
}

function executableNodeBin() {
  const configured = process.env.NODE_BIN ?? "";
  if (configured) {
    try {
      accessSync(configured, constants.X_OK);
      return configured;
    } catch {
      // Fall through to the current Node executable.
    }
  }
  return process.execPath;
}

function testOutputScript() {
  return process.env.TEST_OUTPUT_SCRIPT || path.join(repoRoot, "scripts", "lib", "test-output.mjs");
}

function resultsRoot() {
  const configured = process.env.CARTULARY_TEST_RESULTS_DIR || "";
  if (configured) {
    return path.isAbsolute(configured) ? configured : path.join(repoRoot, configured);
  }
  return path.join(repoRoot, ".cartulary", "test-results");
}

function runID() {
  return process.env.CARTULARY_TEST_RUN_ID || "adhoc";
}

function runWithContext(command, args, { env = process.env } = {}) {
  const result = spawnSync(command, args, {
    cwd: repoRoot,
    env,
    stdio: "inherit",
  });
  if (result.error) {
    throw result.error;
  }
  return result.status ?? 1;
}

function runTargetSummary(target, status, children) {
  const script = testOutputScript();
  const command = script.endsWith(".mjs") ? executableNodeBin() : script;
  const args = script.endsWith(".mjs")
    ? [script, "target-summary", target, status]
    : ["target-summary", target, status];
  if (children.length > 0) {
    args.push("--children", children.join(","));
  }
  return runWithContext(command, args);
}

function targetSummaryStatus(target) {
  const file = path.join(resultsRoot(), runID(), target, "target-summary.json");
  if (!existsSync(file)) {
    return "missing";
  }
  return JSON.parse(readFileSync(file, "utf8")).status ?? "missing";
}

function runMakeTarget(target) {
  const makeBin = process.env.MAKE || "make";
  return runWithContext(makeBin, ["--no-print-directory", target], {
    env: {
      ...process.env,
      CARTULARY_TEST_TARGET: target,
    },
  });
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const slice = phaseSlice(options.phase, {
    serviceBackedOnly: options.mode === "service-backed",
  });
  if (!slice) {
    throw new Error(`unknown phase ${options.phase}; expected one of tools/phase*_test_map.json`);
  }

  const children = slice.child_targets.map((entry) => entry.target);
  if (children.length === 0) {
    process.stdout.write(
      `[NOOP] ${slice.target} phase=${slice.phase} mode=${options.mode} children=0\n`,
    );
    const summaryStatus = runTargetSummary(slice.target, "pass", []);
    if (summaryStatus !== 0) {
      return summaryStatus;
    }
    return targetSummaryStatus(slice.target) === "pass" ? 0 : 1;
  }

  let childStatus = 0;
  for (const child of children) {
    childStatus = runMakeTarget(child);
    if (childStatus !== 0) {
      break;
    }
  }

  const requestedStatus = childStatus === 0 ? "pass" : "fail";
  const summaryStatus = runTargetSummary(slice.target, requestedStatus, children);
  if (childStatus !== 0) {
    return childStatus;
  }
  if (summaryStatus !== 0) {
    return summaryStatus;
  }
  return targetSummaryStatus(slice.target) === "pass" ? 0 : 1;
}

try {
  process.exit(main());
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`phase slice failed: ${message}\n`);
  process.exit(1);
}
