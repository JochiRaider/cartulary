#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import { scanExecutableDocumentationReads } from "./documentation-boundary.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");

try {
  scanExecutableDocumentationReads(repoRoot);
  process.stdout.write("documentation boundary check passed\n");
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exitCode = 11;
}
