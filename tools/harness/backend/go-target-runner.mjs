#!/usr/bin/env node

import { pathToFileURL } from "node:url";

import { runGoTargetCLI } from "./backend-target-execution.mjs";

export { runGoTargetCLI } from "./backend-target-execution.mjs";

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const status = await runGoTargetCLI(process.argv.slice(2));
  process.exit(status);
}
