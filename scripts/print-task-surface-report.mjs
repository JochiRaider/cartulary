#!/usr/bin/env node

import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const checkMode = process.argv.includes("--check");
const jsonMode = process.argv.includes("--json");
const makefilePath = resolvePath(
  process.env.CARTULARY_TASK_SURFACE_MAKEFILE ?? "Makefile",
);
const manifestPath = resolvePath(
  process.env.CARTULARY_TASK_SURFACE_MANIFEST ?? "tools/task_surface_manifest.json",
);

const validClassifications = new Set(["public", "check_internal", "helper_only"]);
const validInclusions = new Set(["test", "check", "ci", "release-check", "helper_only"]);

function main() {
  const makefile = readFileSync(makefilePath, "utf8");
  const manifest = readJSON(manifestPath);
  const phonyTargets = collectPhonyTargets(makefile);
  const helpEntries = collectHelpEntries(makefile);
  const targetBlocks = collectTargetBlocks(makefile, phonyTargets);
  const targetScriptRefs = new Map(
    phonyTargets.map((target) => [target, collectDirectScriptRefs(targetBlocks.get(target) ?? "")]),
  );
  const phaseDependencies = collectPhaseDependencies();
  const errors = validateTaskSurface({
    helpEntries,
    manifest,
    phonyTargets,
    targetScriptRefs,
  });
  const report = buildReport({
    errors,
    helpEntries,
    manifest,
    phaseDependencies,
    phonyTargets,
    targetScriptRefs,
  });

  if (jsonMode) {
    process.stdout.write(`${JSON.stringify(report, null, 2)}\n`);
  } else {
    printHumanReport(report);
  }

  if (checkMode && errors.length > 0) {
    process.exit(1);
  }
}

function resolvePath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function readJSON(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch (error) {
    throw new Error(`failed to read JSON ${file}: ${error.message}`);
  }
}

function collectPhonyTargets(makefile) {
  const targets = [];
  for (const line of makefile.split(/\r?\n/)) {
    if (!line.startsWith(".PHONY:")) {
      continue;
    }
    targets.push(...line.replace(".PHONY:", "").trim().split(/\s+/).filter(Boolean));
  }
  return targets;
}

function collectHelpEntries(makefile) {
  const entries = new Map();
  for (const line of makefile.split(/\r?\n/)) {
    const match = /^\s*'  make ([A-Za-z0-9_.-]+)(?:\s+[^']*)?'/.exec(line);
    if (!match) {
      continue;
    }
    const target = match[1];
    entries.set(target, line.trim().replace(/^'/, "").replace(/' \\?$/, ""));
  }
  return entries;
}

function collectTargetBlocks(makefile, targets) {
  const lines = makefile.split(/\r?\n/);
  const targetSet = new Set(targets);
  const blocks = new Map();

  for (let index = 0; index < lines.length; index += 1) {
    const match = /^([A-Za-z0-9_.-]+):/.exec(lines[index]);
    if (!match || !targetSet.has(match[1]) || lines[index].includes(": export ")) {
      continue;
    }
    const target = match[1];
    const blockLines = [lines[index]];
    for (let next = index + 1; next < lines.length; next += 1) {
      if (/^[^\s#][^:]*:/.test(lines[next])) {
        break;
      }
      blockLines.push(lines[next]);
    }
    blocks.set(target, blockLines.join("\n"));
  }

  return blocks;
}

function collectDirectScriptRefs(source) {
  const refs = new Set();
  for (const match of source.matchAll(/(?:\.\/)?scripts\/[A-Za-z0-9_.\/-]+/g)) {
    refs.add(match[0].replace(/^\.\//, ""));
  }
  return Array.from(refs).sort();
}

function collectPhaseDependencies() {
  const toolsDir = path.join(repoRoot, "tools");
  const rows = [];
  if (!existsSync(toolsDir)) {
    return rows;
  }

  for (const file of readdirSync(toolsDir).sort()) {
    const match = /^(phase\d+)_test_map\.json$/.exec(file);
    if (!match) {
      continue;
    }
    const phase = match[1];
    const manifest = readJSON(path.join(toolsDir, file));
    const counts = new Map();
    for (const section of ["unit", "integration", "e2e"]) {
      for (const entry of manifest[section] ?? []) {
        if (typeof entry.execution_dependency !== "string" || entry.execution_dependency === "") {
          continue;
        }
        const key = `${section}:${entry.execution_dependency}`;
        counts.set(key, (counts.get(key) ?? 0) + 1);
      }
    }
    for (const [key, count] of counts.entries()) {
      const [section, executionDependency] = key.split(":", 2);
      rows.push({ count, execution_dependency: executionDependency, phase, section });
    }
  }
  return rows.sort((left, right) =>
    `${left.phase}:${left.section}:${left.execution_dependency}`.localeCompare(
      `${right.phase}:${right.section}:${right.execution_dependency}`,
    ),
  );
}

function validateTaskSurface({ helpEntries, manifest, phonyTargets, targetScriptRefs }) {
  const errors = [];

  if (manifest.schema_id !== "cartulary.task_surface_manifest.v1") {
    errors.push("tools/task_surface_manifest.json must declare schema_id=cartulary.task_surface_manifest.v1");
  }
  if (!Array.isArray(manifest.targets)) {
    errors.push("tools/task_surface_manifest.json must declare targets[]");
    return errors;
  }

  const phonySet = new Set(phonyTargets);
  const entriesByName = new Map();
  for (const entry of manifest.targets) {
    if (typeof entry.name !== "string" || entry.name.trim() === "") {
      errors.push("task-surface manifest entry has missing name");
      continue;
    }
    if (entriesByName.has(entry.name)) {
      errors.push(`task-surface manifest has duplicate target ${entry.name}`);
      continue;
    }
    entriesByName.set(entry.name, entry);
  }

  for (const target of phonyTargets) {
    if (!entriesByName.has(target)) {
      errors.push(`Makefile .PHONY target ${target} is missing task-surface classification`);
    }
  }
  for (const target of entriesByName.keys()) {
    if (!phonySet.has(target)) {
      errors.push(`task-surface manifest target ${target} is not a Makefile .PHONY target`);
    }
  }

  for (const [target, entry] of entriesByName.entries()) {
    if (!validClassifications.has(entry.classification)) {
      errors.push(`${target} has invalid classification ${JSON.stringify(entry.classification)}`);
    }
    if (!Array.isArray(entry.included_in) || entry.included_in.length === 0) {
      errors.push(`${target} must declare non-empty included_in[]`);
    } else {
      for (const inclusion of entry.included_in) {
        if (!validInclusions.has(inclusion)) {
          errors.push(`${target} has invalid included_in value ${JSON.stringify(inclusion)}`);
        }
      }
    }

    const hasHelp = helpEntries.has(target);
    if (entry.classification === "public" && !hasHelp) {
      errors.push(`public target ${target} is missing a help entry`);
    }
    if (hasHelp && entry.classification !== "public") {
      errors.push(`help entry ${target} must be classified public`);
    }

    const declaredScripts = Array.isArray(entry.backing_scripts) ? entry.backing_scripts : [];
    for (const script of declaredScripts) {
      if (typeof script !== "string" || script.trim() === "") {
        errors.push(`${target} declares an invalid backing script`);
        continue;
      }
      const scriptPath = path.join(repoRoot, script);
      if (!existsSync(scriptPath) || !statSync(scriptPath).isFile()) {
        errors.push(`${target} backing script missing: ${script}`);
      }
    }

    const actualScriptRefs = targetScriptRefs.get(target) ?? [];
    const declaredScriptSet = new Set(declaredScripts);
    for (const script of actualScriptRefs) {
      if (!declaredScriptSet.has(script)) {
        errors.push(`${target} references ${script} but does not declare it in task_surface_manifest.json`);
      }
    }
  }

  for (const target of helpEntries.keys()) {
    if (!phonySet.has(target)) {
      errors.push(`help entry ${target} is not a Makefile .PHONY target`);
    }
  }

  return errors;
}

function buildReport({ errors, helpEntries, manifest, phaseDependencies, phonyTargets, targetScriptRefs }) {
  const entriesByName = new Map((manifest.targets ?? []).map((entry) => [entry.name, entry]));
  const targets = phonyTargets.map((target) => {
    const entry = entriesByName.get(target) ?? {};
    return {
      name: target,
      classification: entry.classification ?? "unclassified",
      has_help: helpEntries.has(target),
      included_in: entry.included_in ?? [],
      backing_scripts: entry.backing_scripts ?? [],
      makefile_script_refs: targetScriptRefs.get(target) ?? [],
    };
  });

  return {
    schema_id: "cartulary.task_surface_report.v1",
    check_passed: errors.length === 0,
    errors,
    targets,
    help_entries: Array.from(helpEntries.keys()).sort(),
    phase_execution_dependencies: phaseDependencies,
  };
}

function printHumanReport(report) {
  console.log("Cartulary task-surface report");
  console.log("");
  console.log("public Make targets:");
  for (const target of report.targets.filter((entry) => entry.classification === "public")) {
    console.log(
      `  ${target.name} help=${target.has_help ? "yes" : "no"} included_in=${target.included_in.join(",")}`,
    );
  }

  console.log("");
  console.log("task classifications:");
  for (const target of report.targets) {
    const scripts = target.backing_scripts.length > 0 ? target.backing_scripts.join(",") : "-";
    console.log(
      `  ${target.name} classification=${target.classification} included_in=${target.included_in.join(",")} scripts=${scripts}`,
    );
  }

  console.log("");
  console.log("phase-map execution dependencies:");
  for (const row of report.phase_execution_dependencies) {
    console.log(
      `  ${row.phase} ${row.section} ${row.execution_dependency} rows=${row.count}`,
    );
  }

  if (report.errors.length > 0) {
    console.log("");
    console.log("drift:");
    for (const error of report.errors) {
      console.log(`  ${error}`);
    }
  }
}

try {
  main();
} catch (error) {
  console.error(`task-surface report failed: ${error.message}`);
  process.exit(1);
}
