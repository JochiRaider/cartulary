#!/usr/bin/env node
import { spawnSync } from "../workspace/child-process.mjs";

import { createRunnerContext, runnerEnv } from "../contract/index.mjs";

function usage() {
  process.stderr.write(`usage:
  cartulary-runner.mjs target-summary <target> [pass|fail] [...]
`);
  process.exit(2);
}

function runTargetSummary(context, args) {
  const command = context.testOutputScript.endsWith(".mjs")
    ? context.nodeBin
    : context.testOutputScript;
  const commandArgs = context.testOutputScript.endsWith(".mjs")
    ? [context.testOutputScript, "target-summary", ...args]
    : ["target-summary", ...args];
  const result = spawnSync(command, commandArgs, {
    cwd: context.repoRoot,
    env: runnerEnv(context),
    stdio: "inherit",
  });
  if (result.error) {
    throw result.error;
  }
  return result.status ?? 1;
}

// Only an executor that owns a target's result calls this: the command entry
// for bootstrap and the operation runner with --report.
function main() {
  const [command, ...rest] = process.argv.slice(2);
  const context = createRunnerContext();
  switch (command) {
    case "target-summary":
      process.exit(runTargetSummary(context, rest));
      break;
    default:
      usage();
  }
}

main();
