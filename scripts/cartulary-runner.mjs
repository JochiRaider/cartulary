#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { createRunnerContext, runnerEnv } from "../tools/harness/core/runner-context.mjs";
import {
  loadSummaryTopologyContext,
  serviceBackedScheduleChildren,
} from "./lib/summary-topology.mjs";
import { publicExitCodeForSummary } from "../tools/harness/core/failure-taxonomy.mjs";

const goTargetCommands = new Set([
  "inspect-aggregate-command",
  "capture-shard",
  "finalize-shards",
  "backend-unit",
  "backend-store",
  "backend-integration",
  "backend-integration-support",
  "backend-process",
]);

function usage() {
  process.stderr.write(`usage:
  cartulary-runner.mjs service-backed-target --target <target> --phase-label <label> --service-wrapper <test-services|none>
  cartulary-runner.mjs summary-target --target <target> --child-target <target> --status <pass|fail> [--phase-label <label>] [--projection <target>]
  cartulary-runner.mjs go-target <target-or-command> [...]
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

function serviceBackedTarget(context, argv) {
  const options = parseFlagArgs(argv);
  const target = options.target || "";
  const phaseLabel = options.phase_label || "";
  const serviceWrapper = options.service_wrapper || "";
  if (
    !target ||
    !phaseLabel ||
    !serviceWrapper ||
    options.projection !== undefined
  ) {
    usage();
  }

  const env = runnerEnv(context, {
    CARTULARY_TEST_GO_TARGET_RUNNER:
      process.env.CARTULARY_TEST_GO_TARGET_RUNNER || context.runnerScript,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  });
  const schedulerArgs = [
    context.runPhaseScript,
    phaseLabel,
    "--",
    context.nodeBin,
    context.serviceBackedScheduleScript,
    "--target",
    target,
    "--manifest",
    context.schedulerManifest,
    "--defer-summary",
  ];

  let status = 0;
  if (serviceWrapper === "test-services") {
    if (!context.testServicesBin) {
      process.stderr.write(
        "TEST_SERVICES_BIN is required for --service-wrapper test-services\n",
      );
      return 2;
    }
    status = runWithContext(
      context,
      context.testServicesBin,
      ["run", "--", ...schedulerArgs],
      {
        env,
      },
    );
  } else if (serviceWrapper === "none") {
    status = runWithContext(context, schedulerArgs[0], schedulerArgs.slice(1), {
      env,
    });
  } else {
    usage();
  }

  const requested = status === 0 ? "pass" : "fail";
  const topologyContext = loadSummaryTopologyContext({
    schedulerManifestPath: context.schedulerManifest,
  });
  const children = serviceBackedScheduleChildren(topologyContext, target);
  if (children.length === 0) {
    process.stderr.write(
      `service-backed schedule ${target} has no derived summary children\n`,
    );
    return 2;
  }
  const summaryArgs = [target, requested, "--children", children.join(",")];
  if (process.env.CARTULARY_SUPPRESS_CHILD_SUCCESS === "1") {
    summaryArgs.push("--quiet-success");
  }
  const summaryStatus = runTargetSummary(context, summaryArgs);
  return summaryStatus === 0
    ? targetPublicExitCode(context, target, status)
    : targetPublicExitCode(context, target, summaryStatus);
}

function summaryTarget(context, argv) {
  const options = parseFlagArgs(argv);
  const target = options.target || "";
  const childTarget = options.child_target || "";
  const requestedStatus = options.status || "";
  const projection = options.projection || "";
  const phaseLabel = options.phase_label || `${target} child ${childTarget}`;
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
    context.runPhaseScript,
    [phaseLabel, "--", context.makeBin, "--no-print-directory", childTarget],
    {
      env: runnerEnv(context, {
        CARTULARY_TEST_TARGET: target,
        CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
      }),
    },
  );
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

function goTarget(context, argv) {
  if (argv.length === 0) {
    usage();
  }
  const command = context.runGoTargetScript.endsWith(".mjs")
    ? context.nodeBin
    : context.runGoTargetScript;
  const args = context.runGoTargetScript.endsWith(".mjs")
    ? [context.runGoTargetScript, ...argv]
    : argv;
  return runWithContext(context, command, args, {
    env: runnerEnv(context),
  });
}

function main() {
  const [command, ...rest] = process.argv.slice(2);
  const context = createRunnerContext();
  if (goTargetCommands.has(command)) {
    process.exit(goTarget(context, [command, ...rest]));
  }
  switch (command) {
    case "service-backed-target":
      process.exit(serviceBackedTarget(context, rest));
      break;
    case "summary-target":
      process.exit(summaryTarget(context, rest));
      break;
    case "go-target":
      process.exit(goTarget(context, rest));
      break;
    case "target-summary":
      process.exit(runTargetSummary(context, rest));
      break;
    default:
      usage();
  }
}

main();
