#!/usr/bin/env node

import { chmodSync, existsSync, readFileSync } from "node:fs";
import path from "node:path";

import {
  prettyJSONString,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  assertBaselineClosure,
  baselineSchemaID,
  buildFromManifest,
  performanceRoster,
} from "./canonical-performance.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const baselineFile = path.join(root, "tools/harness_public_target_duration_baselines.v3.json");

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function parse(argv) {
  const modes = new Set(["publish", "check", "coverage", "observe"]);
  const options = { mode: "", evidence: "", run: "", target: "", view: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (modes.has(arg) && !options.mode) options.mode = arg;
    else if (arg === "--evidence-roots-file") options.evidence = argv[++index] ?? "";
    else if (arg === "--results-dir") options.run = argv[++index] ?? "";
    else if (arg === "--target") options.target = argv[++index] ?? "";
    else if (arg === "--view") options.view = argv[++index] ?? "";
    else throw new Error("usage: canonical-performance-cli.mjs publish|check|coverage|observe [options]");
  }
  return options;
}

function publish(bytes) {
  if (existsSync(baselineFile) && readFileSync(baselineFile, "utf8") === bytes) {
    chmodSync(baselineFile, 0o644);
    return "unchanged";
  }
  secureWriteFile(baselineFile, bytes, { allowedRoot: root });
  chmodSync(baselineFile, 0o644);
  return "published";
}

function observation(runRoot, target) {
  const summary = readJSON(path.join(runRoot, "run-summary.json"));
  validateSchemaSync("cartulary.harness_run_summary.v1", summary);
  if (!target) {
    return `[OBSERVATION] target=${summary.target} duration_ms=${summary.wall_duration_ms} critical_path_ms=${summary.actual_dependency_critical_path_ms}\n`;
  }
  const targetSummary = readJSON(path.join(runRoot, "target-summaries", `${target}.json`));
  validateSchemaSync("cartulary.harness_target_summary.v1", targetSummary);
  return `[OBSERVATION] target=${target} duration_ms=${targetSummary.inclusive_wall_ms} critical_path_ms=${targetSummary.actual_dependency_critical_path_ms}\n`;
}

try {
  const options = parse(process.argv.slice(2));
  const surface = readJSON(path.join(root, "tools/task_surface_owner.json"));
  if (options.mode === "publish") {
    if (!options.evidence) throw new Error("publish requires --evidence-roots-file");
    const built = await buildFromManifest(root, path.resolve(options.evidence), surface);
    if (built.manifest.mode !== "baseline") throw new Error("baseline publication requires mode=baseline evidence");
    const disposition = publish(prettyJSONString(built.reference));
    process.stdout.write(`[BASELINE] status=${disposition} targets=${built.reference.targets.length}\n`);
  } else if (options.mode === "check") {
    if (!options.evidence) throw new Error("check requires --evidence-roots-file");
    const built = await buildFromManifest(root, path.resolve(options.evidence), surface);
    if (built.manifest.mode !== "comparison") throw new Error("performance check requires mode=comparison evidence");
    if (built.comparison.failures.length > 0) {
      throw new Error(`performance acceptance failed ${built.comparison.failures.join(",")}`);
    }
    process.stdout.write(`[PERFORMANCE] status=pass targets=${built.comparison.rows.length}\n`);
  } else if (options.mode === "coverage") {
    const roster = performanceRoster(surface);
    if (!existsSync(baselineFile)) {
      process.stdout.write(`[BASELINE] status=pending-publication targets=${roster.size}\n`);
    } else {
      const baseline = readJSON(baselineFile);
      assertBaselineClosure(baseline, surface);
      process.stdout.write(`[BASELINE] status=covered targets=${baseline.targets.length}\n`);
    }
  } else if (options.mode === "observe") {
    if (!options.run) throw new Error("observe requires --results-dir");
    process.stdout.write(observation(path.resolve(options.run), options.target));
  } else {
    throw new Error("unknown canonical performance mode");
  }
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
}
