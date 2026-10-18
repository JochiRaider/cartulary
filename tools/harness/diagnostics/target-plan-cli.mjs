#!/usr/bin/env node
import path from "node:path";
import { validateSchemaSync } from "../contract/index.mjs";
import { buildTargetPlan, readPlanInputs, renderTargetPlan } from "./target-plan.mjs";
const root = path.resolve(import.meta.dirname, "../../..");
function parseArgs(argv) {
  const options = { target: "", detail: "summary", json: false, inputs: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--json") options.json = true;
    else if (arg === "--view") options.detail = argv[++index] ?? "";
    else if (arg === "--target") options.target = argv[++index] ?? "";
    else if (arg === "--inputs") options.inputs = argv[++index] ?? "";
    else throw new Error("invalid target-plan argument");
  }
  if (!options.target) throw new Error("TARGET is required");
  if (!["summary", "graph", "rows", "artifacts"].includes(options.detail)) throw new Error("invalid DETAIL");
  return options;
}
try {
  const options = parseArgs(process.argv.slice(2));
  const plan = buildTargetPlan(root, options.target, readPlanInputs(root, options.inputs));
  validateSchemaSync(plan.schema_id, plan);
  process.stdout.write(`${options.json ? JSON.stringify(plan, null, 2) : renderTargetPlan(plan, options.detail)}\n`);
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 2;
}
