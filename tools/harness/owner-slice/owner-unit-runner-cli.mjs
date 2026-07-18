#!/usr/bin/env node

import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { redactValue, validateSchemaSync } from "../contract/index.mjs";
import { executeOwnerSliceUnit } from "./execution.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function options(argv) {
  const parsed = { plan: "", unit: "", result: "", artifactRoot: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--plan") parsed.plan = argv[++index] ?? "";
    else if (arg === "--unit") parsed.unit = argv[++index] ?? "";
    else if (arg === "--result") parsed.result = argv[++index] ?? "";
    else if (arg === "--artifact-root") parsed.artifactRoot = argv[++index] ?? "";
    else throw new Error("usage: owner-unit-runner-cli.mjs --plan <path> --unit <id> --result <path> --artifact-root <path>");
  }
  if (!parsed.plan || !parsed.unit || !parsed.result || !parsed.artifactRoot) {
    throw new Error("usage: owner-unit-runner-cli.mjs --plan <path> --unit <id> --result <path> --artifact-root <path>");
  }
  return parsed;
}

try {
  const selected = options(process.argv.slice(2));
  const plan = JSON.parse(readFileSync(path.resolve(selected.plan), "utf8"));
  validateSchemaSync(plan.schema_id, plan);
  const execution = executeOwnerSliceUnit(root, plan, selected.unit, {
    artifactRoot: path.resolve(selected.artifactRoot),
  });
  writeFileSync(path.resolve(selected.result), `${JSON.stringify(redactValue(execution), null, 2)}\n`, "utf8");
  process.exitCode = execution.status === "pass"
    ? 0
    : execution.row_results.some((row) => row.terminal_state === "failed")
      ? 10
      : execution.row_results.find((row) => row.exit_code !== 0)?.exit_code ?? 11;
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 11;
}
