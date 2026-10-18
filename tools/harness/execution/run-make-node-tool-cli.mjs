#!/usr/bin/env node
import { spawnSync } from "../workspace/child-process.mjs";
import { buildMakeNodeToolChildEnv, buildMakeNodeToolInvocation, makeNodeToolNames, UsageError } from "../command-surface/make-node-tools.mjs";

// Public read-only commands own their output. Scheduled operations are captured
// once by the native runtime or canonical row adapter, never a second wrapper.
function usage() {
  process.stderr.write(`usage: run-make-node-tool.mjs <${makeNodeToolNames().join("|")}>\n`);
}

function main(argv) {
  const [target, ...extra] = argv;
  if (!target || extra.length > 0) {
    usage();
    return 2;
  }

  let invocation;
  try {
    invocation = buildMakeNodeToolInvocation(target, process.env);
  } catch (error) {
    if (error instanceof UsageError) {
      process.stderr.write(`${error.usage ?? error.message}\n`);
      return 2;
    }
    throw error;
  }

  const child = spawnSync(process.execPath, [invocation.script, ...invocation.args], {
    env: buildMakeNodeToolChildEnv(target, process.env),
    stdio: "inherit",
  });
  if (child.error) {
    throw child.error;
  }
  return child.status ?? 1;
}

try { process.exitCode = main(process.argv.slice(2)); }
catch (error) { process.stderr.write(`make node tool failed: ${error.message}\n`); process.exitCode = 1; }
