#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import {
  createRunnerContext,
  publicExitCodeForSummary,
  runnerEnv,
} from "../contract/index.mjs";
function usage() {
  process.stderr.write(`usage:
  cartulary-runner.mjs summary-target --target <target> --child-target <target> --status <pass|fail> [--step-label <label>] [--projection <target>]
  cartulary-runner.mjs target-summary <target> [pass|fail] [...]
`);
  process.exit(2);
}

function parseFlagArgs(argv) {
  const options = {};
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (!arg.startsWith("--")) {
      usage();
    }
    const value = argv[index + 1];
    if (value === undefined) {
      usage();
    }
    options[arg.slice(2).replaceAll("-", "_")] = value;
    index += 1;
  }
  return options;
}

function runWithContext(
  context,
  command,
  args,
  { env = process.env, stdio = "inherit" } = {},
) {
  const result = spawnSync(command, args, {
    cwd: context.repoRoot,
    env,
    stdio,
  });
  if (result.error) {
    throw result.error;
  }
  return result.status ?? 1;
}

function runTargetSummary(context, args) {
  const command = context.testOutputScript.endsWith(".mjs")
    ? context.nodeBin
    : context.testOutputScript;
  const commandArgs = context.testOutputScript.endsWith(".mjs")
    ? [context.testOutputScript, "target-summary", ...args]
    : ["target-summary", ...args];
  return runWithContext(context, command, commandArgs, {
    env: runnerEnv(context),
  });
}

function targetPublicExitCode(context, target, fallbackStatus) {
  const summaryFile = path.join(context.resultsDir, context.runId, target, "tool-run-summary.json");
  if (!existsSync(summaryFile)) {
    return fallbackStatus === 0 ? 0 : 1;
  }
  try {
    const summary = JSON.parse(readFileSync(summaryFile, "utf8"));
    return publicExitCodeForSummary(summary);
  } catch {
    return fallbackStatus === 0 ? 0 : 1;
  }
}

function summaryTarget(context, argv) {
  const options = parseFlagArgs(argv);
  const target = options.target || "";
  const childTarget = options.child_target || "";
  const requestedStatus = options.status || "";
  const projection = options.projection || "";
  const stepLabel = options.step_label || `${target} child ${childTarget}`;
  if (
    !target ||
    !childTarget ||
    !["pass", "fail"].includes(requestedStatus) ||
    (projection !== "" && projection.startsWith("--"))
  ) {
    usage();
  }

  const childStatus = runWithContext(
    context,
    context.runStepScript,
    [stepLabel, "--", context.makeBin, "--no-print-directory", childTarget],
    {
      env: runnerEnv(context, {
        CARTULARY_TEST_TARGET: target,
        CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
      }),
    },
  );
  if (process.env.CARTULARY_HARNESS_GRAPH_CHILD === "1") {
    return childStatus;
  }
  const summaryArgs = [target, childStatus === 0 ? requestedStatus : "fail"];
  if (projection) {
    summaryArgs.push("--projection", projection);
    summaryArgs.push("--skipped-from-child", childTarget);
  }
  const summaryStatus = runTargetSummary(context, summaryArgs);
  return summaryStatus === 0
    ? targetPublicExitCode(context, target, childStatus)
    : targetPublicExitCode(context, target, summaryStatus);
}

function main() {
  const [command, ...rest] = process.argv.slice(2);
  const context = createRunnerContext();
  switch (command) {
    case "summary-target":
      process.exit(summaryTarget(context, rest));
      break;
    case "target-summary":
      if (process.env.CARTULARY_HARNESS_GRAPH_CHILD === "1") process.exit(0);
      process.exit(runTargetSummary(context, rest));
      break;
    default:
      usage();
  }
}

main();
