#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadTestCatalog } from "../test-catalog/index.mjs";
import { reportCommandFailure } from "../runtime/command-failure.mjs";
import { publicExitCodeForFailure } from "../contract/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "..", "..", "..");
try {
  process.stdout.write(`${JSON.stringify(loadTestCatalog(root).summary)}\n`);
} catch (error) {
  const failure = reportCommandFailure(root, error, { failure_class: "config", failure_reason: "configuration_error" });
  process.stderr.write(`test catalog configuration rejected: ${error instanceof Error ? error.message : "invalid catalog"}\n`);
  process.exitCode = publicExitCodeForFailure(failure);
}
