#!/usr/bin/env node

import { spawnSync } from "node:child_process";

import {
  buildMakeNodeToolChildEnv,
  buildMakeNodeToolInvocation,
  makeNodeToolNames,
  UsageError,
} from "./lib/make-node-tools.mjs";

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

try {
  process.exit(main(process.argv.slice(2)));
} catch (error) {
  process.stderr.write(`make node tool failed: ${error.message}\n`);
  process.exit(1);
}
