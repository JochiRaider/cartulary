#!/usr/bin/env node

import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../contract/index.mjs";
import { buildOwnerTaskGuide } from "./owner-diagnostics.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function parseArgs(argv) {
  const options = { mode: "", ownerID: "", detail: "summary", jsonValue: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--mode") options.mode = argv[++index] ?? "";
    else if (arg === "--owner") options.ownerID = argv[++index] ?? "";
    else if (arg === "--detail") options.detail = argv[++index] ?? "";
    else if (arg === "--json") options.jsonValue = "1";
    else if (arg === "--json-value") options.jsonValue = argv[++index] ?? "";
    else throw new Error("invalid owner diagnostics usage");
  }
  if (options.mode !== "task-guide") throw new Error("invalid owner diagnostics mode");
  if (options.jsonValue !== "" && options.jsonValue !== "1") throw new Error("JSON accepts only exact 1");
  if (options.jsonValue === "1" && process.env.CARTULARY_OUTPUT_MODE === "machine") {
    throw new Error("JSON=1 cannot be combined with CARTULARY_OUTPUT_MODE=machine");
  }
  if (!["summary", "rows", "coverage"].includes(options.detail)) throw new Error("invalid DETAIL");
  return options;
}

function humanTaskGuide(summary, detail) {
  if (detail === "rows") return `${summary.rows.map((row) => `${row.row_id} runner=${row.runner} evidence=${row.evidence_class} tier=${row.minimum_tier} fixture=${row.fixture_capability}\n  ${row.command}`).join("\n")}\n`;
  if (detail === "coverage") return `${JSON.stringify(summary.coverage, null, 2)}\n${summary.focused_commands.join("\n")}\n`;
  return `${[
    `owner ${summary.owner_id}`,
    ...summary.focused_commands.map((command) => `focused ${command}`),
    ...summary.generated_commands.map((command) => `generated ${command}`),
    ...summary.broader_commands.map((command) => `broader ${command}`),
    ...(summary.release_gate ? [`release ${summary.release_gate}`] : []),
  ].join("\n")}\n`;
}

try {
  const options = parseArgs(process.argv.slice(2));
  const summary = buildOwnerTaskGuide(root, options.ownerID);
  validateSchemaSync(summary.schema_id, summary);
  if (options.jsonValue === "1") process.stdout.write(`${JSON.stringify(summary)}\n`);
  else process.stdout.write(humanTaskGuide(summary, options.detail));
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 2;
}
