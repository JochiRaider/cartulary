#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function parse(argv) {
  const options = { target: "", groups: [], groupTargets: new Map(), children: [] };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") options.target = argv[++index] ?? "";
    else if (arg === "--groups") options.groups = (argv[++index] ?? "").split(",").filter(Boolean);
    else if (arg === "--group-targets") {
      for (const entry of (argv[++index] ?? "").split(",").filter(Boolean)) {
        const separator = entry.lastIndexOf("=");
        if (separator < 1) throw new Error("invalid browser group target mapping");
        options.groupTargets.set(entry.slice(0, separator), entry.slice(separator + 1));
      }
    }
    else if (arg === "--children") options.children = (argv[++index] ?? "").split(",").filter(Boolean);
    else throw new Error("invalid browser target finalizer arguments");
  }
  if (!options.target || options.groups.length === 0 || options.groupTargets.size !== options.groups.length) {
    throw new Error("browser target finalizer requires target and exact group targets");
  }
  return options;
}

function runRoot() {
  const results = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!results || !runID) throw new Error("browser target finalizer requires result-root identity");
  return path.resolve(root, results, runID);
}

function groupPassed(base, groupID, target) {
  const file = path.join(base, target, "browser-groups", groupID, "browser-group-result.json");
  return existsSync(file) && JSON.parse(readFileSync(file, "utf8")).status === "pass";
}

const options = parse(process.argv.slice(2));
const requested = options.groups.every((group) => groupPassed(runRoot(), group, options.groupTargets.get(group)))
  ? "pass"
  : "fail";
const helper = process.env.TEST_OUTPUT_SCRIPT || path.join(root, "tools", "harness", "output", "test-output.mjs");
const node = process.env.NODE_BIN || process.execPath;
const args = [helper, "target-summary", options.target, requested];
if (options.children.length > 0) args.push("--children", options.children.join(","));
const child = spawnSync(helper.endsWith(".mjs") ? node : helper, helper.endsWith(".mjs") ? args : args.slice(1), {
  cwd: root,
  env: process.env,
  stdio: "inherit",
});
process.exitCode = child.status ?? 11;
