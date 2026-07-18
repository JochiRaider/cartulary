#!/usr/bin/env node

import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../contract/index.mjs";
import { buildModuleAuthorTaskGuide, explainTestOwner } from "./owner-diagnostics.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function parseArgs(argv) {
  const options = { mode: "", ownerID: "", role: "", jsonValue: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--mode") options.mode = argv[++index] ?? "";
    else if (arg === "--owner") options.ownerID = argv[++index] ?? "";
    else if (arg === "--role") options.role = argv[++index] ?? "";
    else if (arg === "--json") options.jsonValue = "1";
    else if (arg === "--json-value") options.jsonValue = argv[++index] ?? "";
    else throw new Error("invalid owner diagnostics usage");
  }
  if (!new Set(["explain", "task-guide"]).has(options.mode)) throw new Error("invalid owner diagnostics mode");
  if (options.jsonValue !== "" && options.jsonValue !== "1") throw new Error("JSON accepts only exact 1");
  if (options.jsonValue === "1" && process.env.CARTULARY_OUTPUT_MODE === "machine") {
    throw new Error("JSON=1 cannot be combined with CARTULARY_OUTPUT_MODE=machine");
  }
  return options;
}

function humanExplanation(summary) {
  const lines = [
    `owner ${summary.owner_id}`,
    `manifest ${summary.manifest_path}`,
    `rows ${summary.row_count} service-backed ${summary.service_backed_row_count}`,
    `families ${summary.families.map((entry) => `${entry.family_id}:${entry.row_count}`).join(", ")}`,
    `runners ${Object.entries(summary.runner_counts).map(([key, count]) => `${key}:${count}`).join(", ")}`,
    `evidence ${Object.entries(summary.evidence_counts).map(([key, count]) => `${key}:${count}`).join(", ")}`,
    `focused ${summary.commands.full_owner}`,
  ];
  if (summary.commands.service_backed) lines.push(`service-backed ${summary.commands.service_backed}`);
  return `${lines.join("\n")}\n`;
}

function humanTaskGuide(summary) {
  return `${[
    `role ${summary.role}`,
    `owner ${summary.owner_id}`,
    ...summary.focused_commands.map((command) => `focused ${command}`),
    ...summary.generated_commands.map((command) => `generated ${command}`),
    ...summary.broader_commands.map((command) => `broader ${command}`),
    ...(summary.release_gate ? [`release ${summary.release_gate}`] : []),
  ].join("\n")}\n`;
}

try {
  const options = parseArgs(process.argv.slice(2));
  const summary = options.mode === "explain"
    ? explainTestOwner(root, options.ownerID)
    : buildModuleAuthorTaskGuide(root, options.ownerID, options.role);
  validateSchemaSync(summary.schema_id, summary);
  if (options.jsonValue === "1") process.stdout.write(`${JSON.stringify(summary)}\n`);
  else process.stdout.write(options.mode === "explain" ? humanExplanation(summary) : humanTaskGuide(summary));
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 2;
}
