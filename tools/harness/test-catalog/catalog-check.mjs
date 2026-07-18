#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadTestCatalog } from "./test-catalog.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "..", "..", "..");
process.stdout.write(`${JSON.stringify(loadTestCatalog(root).summary)}\n`);
