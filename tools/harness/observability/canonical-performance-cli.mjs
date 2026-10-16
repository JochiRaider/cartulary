#!/usr/bin/env node

import { chmodSync, existsSync } from "node:fs";
import path from "node:path";
import { parseStrictJSON } from "../contract/strict-json.mjs";
import { validateQualification } from "./qualification/validator.mjs";
import { manifestSchema as qualificationSchema } from "./qualification/policy.mjs";

import {
  prettyJSONString,
  readRetainedArtifactBytes,
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
  return parseStrictJSON(readRetainedArtifactBytes(path.dirname(file), path.basename(file), 16 * 1024 * 1024).toString("utf8"));
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
  if (existsSync(baselineFile) && readRetainedArtifactBytes(path.dirname(baselineFile), path.basename(baselineFile), 16 * 1024 * 1024).toString("utf8") === bytes) {
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
    const selected = readJSON(path.resolve(options.evidence));
    if (selected.schema_id === qualificationSchema) {
      const result = await validateQualification(path.resolve(options.evidence));
      for (const workload of result.workloads) for (const cell of workload.cells) {
        const n = (value) => Number(value.toPrecision(5));
        const stats = cell.stats ? ` cpu_mean=${n(cell.stats.cpu_mean_percent_one_core)}% cpu_mad=${n(cell.stats.cpu_mad_percent_one_core)}% cpu_error_bound=${n(cell.stats.cpu_error_bound_percent_one_core)}% cpu_upper_diagnostic=${n(cell.stats.cpu_upper_percent_one_core)}% rss_upper_mib=${n(cell.stats.rss_upper_bytes / 1048576)} final_ms=${n(cell.stats.finalization_upper_ms)}` : "";
        process.stdout.write(`[QUALIFICATION-CELL] workload=${workload.workload} mode=${cell.mode} readers=${cell.readers} eligible=${cell.eligible} passed=${cell.passed}${stats}\n`);
      }
      if (!result.passed) throw Object.assign(Error(`instrumentation qualification failed ${result.failures.join(",")}`), { exit_code: 13 });
      process.stdout.write(`[QUALIFICATION] status=pass workloads=${result.workloads.length}\n`);
    } else {
    const built = await buildFromManifest(root, path.resolve(options.evidence), surface);
    if (built.manifest.mode !== "comparison") throw new Error("performance check requires mode=comparison evidence");
    if (built.comparison.failures.length > 0) {
      throw Object.assign(new Error(`performance acceptance failed ${built.comparison.failures.join(",")}`), { exit_code: 13 });
    }
    process.stdout.write(`[PERFORMANCE] status=pass targets=${built.comparison.rows.length}\n`);
    }
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
  process.exitCode = error.exit_code ?? 1;
}
