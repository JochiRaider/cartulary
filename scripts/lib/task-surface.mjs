import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");
export const defaultTaskSurfaceManifestPath = path.join(
  repoRoot,
  "tools",
  "task_surface_manifest.json",
);
export const defaultGeneratedMakePath = path.join(repoRoot, "tools", "task_surface.generated.mk");
export const taskSurfaceSchemaID = "cartulary.task_surface_manifest.v4";

const validClassifications = new Set(["public", "check_internal", "helper_only"]);
const validInclusions = new Set(["test", "check", "ci", "release-check", "helper_only"]);
const defaultHelpTierName = "daily";
const defaultHelpMaxEntries = 20;

export function resolveRepoPath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

export function readJSON(file) {
  return JSON.parse(readFileSync(resolveRepoPath(file), "utf8"));
}

export function loadTaskSurfaceManifest(file = defaultTaskSurfaceManifestPath) {
  const manifestPath = resolveRepoPath(file);
  const manifest = readJSON(manifestPath);
  validateTaskSurfaceManifest(manifest, manifestPath);
  return { manifest, manifestPath };
}

export function targetEntries(manifest) {
  return manifest.targets ?? [];
}

export function targetEntryMap(manifest) {
  return new Map(targetEntries(manifest).map((entry) => [entry.name, entry]));
}

export function harnessCheckEntries(manifest) {
  return manifest.harness_checks ?? [];
}

export function harnessCheckEntryMap(manifest) {
  return new Map(harnessCheckEntries(manifest).map((entry) => [entry.name, entry]));
}

export function helpTiers(manifest) {
  return manifest.help_tiers ?? [];
}

export function summaryEntryMap(manifest) {
  return new Map([...targetEntryMap(manifest), ...harnessCheckEntryMap(manifest)]);
}

export function summaryProfile(manifest, name) {
  const profile = manifest.summary_profiles?.[name];
  if (!profile) {
    throw new Error(`unknown task-surface summary profile ${name}`);
  }
  return {
    name,
    targets: [...profile.targets],
    groups: (profile.groups ?? []).map((group) => ({
      name: group.name,
      targets: [...group.targets],
    })),
  };
}

export function summaryProfileArgs(manifest, name) {
  const profile = summaryProfile(manifest, name);
  const args = {
    targets: profile.targets,
    groupsSpec: profile.groups
      .map((group) => `${group.name}=${group.targets.join(",")}`)
      .join(";"),
  };
  return args;
}

export function projectionChildren(manifest, target) {
  const entry = summaryEntryMap(manifest).get(target);
  const children = entry?.summary_projection?.children ?? [];
  return [...children];
}

export function harnessTierChecks(manifest, name) {
  const tier = manifest.harness_tiers?.[name];
  if (!tier) {
    throw new Error(`unknown harness tier ${name}`);
  }
  return [...tier.checks];
}

export function harnessCheck(manifest, name) {
  const check = harnessCheckEntryMap(manifest).get(name);
  if (!check) {
    throw new Error(`unknown harness check ${name}`);
  }
  return {
    name: check.name,
    backing_scripts: [...check.backing_scripts],
  };
}

export function makeIdentifier(value) {
  return value.replace(/[^A-Za-z0-9_]/g, "_").toUpperCase();
}

function validateTaskSurfaceManifest(manifest, manifestPath) {
  const errors = collectTaskSurfaceManifestErrors(manifest);
  if (errors.length > 0) {
    throw new Error(`${manifestPath} is invalid:\n${errors.map((error) => `  - ${error}`).join("\n")}`);
  }
}

export function collectTaskSurfaceManifestErrors(manifest) {
  const errors = [];
  if (manifest.schema_id !== taskSurfaceSchemaID) {
    errors.push(`schema_id must be ${taskSurfaceSchemaID}`);
  }
  if (!Array.isArray(manifest.targets) || manifest.targets.length === 0) {
    errors.push("targets[] must be a non-empty array");
    return errors;
  }

  const targets = new Map();
  for (const [index, entry] of manifest.targets.entries()) {
    const label = `targets[${index + 1}]`;
    if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    if (typeof entry.name !== "string" || entry.name.trim() === "") {
      errors.push(`${label}.name must be a non-empty string`);
      continue;
    }
    if (targets.has(entry.name)) {
      errors.push(`duplicate target ${entry.name}`);
      continue;
    }
    targets.set(entry.name, entry);
    if (!validClassifications.has(entry.classification)) {
      errors.push(`${entry.name} has invalid classification ${JSON.stringify(entry.classification)}`);
    }
    if (!Array.isArray(entry.included_in) || entry.included_in.length === 0) {
      errors.push(`${entry.name} must declare included_in[]`);
    } else {
      for (const inclusion of entry.included_in) {
        if (!validInclusions.has(inclusion)) {
          errors.push(`${entry.name} has invalid included_in value ${JSON.stringify(inclusion)}`);
        }
      }
    }
    if (entry.backing_scripts !== undefined) {
      if (!Array.isArray(entry.backing_scripts)) {
        errors.push(`${entry.name}.backing_scripts must be an array`);
      } else {
        for (const script of entry.backing_scripts) {
          if (typeof script !== "string" || script.trim() === "") {
            errors.push(`${entry.name} declares an invalid backing script`);
          } else if (!existsSync(path.join(repoRoot, script))) {
            errors.push(`${entry.name} backing script missing: ${script}`);
          }
        }
      }
    }
    if (entry.summary_projection !== undefined) {
      const children = entry.summary_projection?.children;
      if (!Array.isArray(children)) {
        errors.push(`${entry.name}.summary_projection.children must be an array`);
      }
    }
  }

  const harnessChecks = new Map();
  if (!Array.isArray(manifest.harness_checks)) {
    errors.push("harness_checks[] must be an array");
  } else {
    for (const [index, entry] of manifest.harness_checks.entries()) {
      const label = `harness_checks[${index + 1}]`;
      if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
        errors.push(`${label} must be an object`);
        continue;
      }
      if (typeof entry.name !== "string" || entry.name.trim() === "") {
        errors.push(`${label}.name must be a non-empty string`);
        continue;
      }
      if (targets.has(entry.name)) {
        errors.push(`harness check ${entry.name} conflicts with a Make target`);
        continue;
      }
      if (harnessChecks.has(entry.name)) {
        errors.push(`duplicate harness check ${entry.name}`);
        continue;
      }
      harnessChecks.set(entry.name, entry);
      if (!Array.isArray(entry.backing_scripts) || entry.backing_scripts.length === 0) {
        errors.push(`${entry.name}.backing_scripts must be a non-empty array`);
      } else {
        for (const script of entry.backing_scripts) {
          if (typeof script !== "string" || script.trim() === "") {
            errors.push(`${entry.name} declares an invalid backing script`);
          } else if (!existsSync(path.join(repoRoot, script))) {
            errors.push(`${entry.name} backing script missing: ${script}`);
          }
        }
      }
      if (entry.summary_projection !== undefined) {
        const children = entry.summary_projection?.children;
        if (!Array.isArray(children)) {
          errors.push(`${entry.name}.summary_projection.children must be an array`);
        }
      }
    }
  }

  const summaryEntries = new Map([...targets, ...harnessChecks]);
  for (const entry of summaryEntries.values()) {
    for (const child of entry.summary_projection?.children ?? []) {
      if (!summaryEntries.has(child)) {
        errors.push(`${entry.name} summary projection references unknown child target ${child}`);
      }
    }
  }

  validateHelpTiers(errors, targets, manifest.help_tiers);

  validateNamedTargetLists(errors, summaryEntries, manifest.summary_profiles, "summary_profiles");
  validateSummaryProfileAccountingRoots(errors, summaryEntries, manifest.summary_profiles);
  validateHarnessTiers(errors, harnessChecks, manifest.harness_tiers);

  if (!manifest.sequences || typeof manifest.sequences !== "object" || Array.isArray(manifest.sequences)) {
    errors.push("sequences must be an object");
  } else {
    for (const [name, sequence] of Object.entries(manifest.sequences)) {
      if (!targets.has(name)) {
        errors.push(`sequence ${name} does not match a declared target`);
      }
      if (typeof sequence.summary_profile !== "string" || !manifest.summary_profiles?.[sequence.summary_profile]) {
        errors.push(`sequence ${name} references unknown summary profile ${JSON.stringify(sequence.summary_profile)}`);
      }
      if (!Array.isArray(sequence.steps) || sequence.steps.length === 0) {
        errors.push(`sequence ${name} must declare steps[]`);
        continue;
      }
      for (const [index, step] of sequence.steps.entries()) {
        const label = `sequence ${name} steps[${index + 1}]`;
        if (!["step", "parallel"].includes(step?.type)) {
          errors.push(`${label}.type must be step or parallel`);
        }
        if (typeof step?.target !== "string" || !targets.has(step.target)) {
          errors.push(`${label}.target references unknown target ${JSON.stringify(step?.target)}`);
        }
        if (step.type === "parallel" && typeof step.jobs_variable !== "string" && !Number.isInteger(step.jobs)) {
          errors.push(`${label} parallel step must declare jobs or jobs_variable`);
        }
      }
    }
  }

  return errors;
}

function projectedChildrenForTarget(targets, targetName, seen = new Set()) {
  if (seen.has(targetName)) {
    return new Set();
  }
  seen.add(targetName);
  const entry = targets.get(targetName);
  const projected = new Set();
  for (const child of entry?.summary_projection?.children ?? []) {
    projected.add(child);
    for (const descendant of projectedChildrenForTarget(targets, child, seen)) {
      projected.add(descendant);
    }
  }
  return projected;
}

function validateSummaryProfileAccountingRoots(errors, targets, summaryProfiles) {
  if (!summaryProfiles || typeof summaryProfiles !== "object" || Array.isArray(summaryProfiles)) {
    return;
  }

  for (const [name, profile] of Object.entries(summaryProfiles)) {
    if (!Array.isArray(profile?.targets)) {
      continue;
    }
    const targetSet = new Set(profile.targets);
    for (const target of profile.targets) {
      for (const child of projectedChildrenForTarget(targets, target)) {
        if (targetSet.has(child)) {
          errors.push(
            `summary_profiles.${name}.targets must not include both aggregate target ${target} and projected child ${child}`,
          );
        }
      }
    }
  }
}

function validateNamedTargetLists(errors, targets, collection, label) {
  if (!collection || typeof collection !== "object" || Array.isArray(collection)) {
    errors.push(`${label} must be an object`);
    return;
  }
  for (const [name, value] of Object.entries(collection)) {
    const targetList = value?.targets;
    if (!Array.isArray(targetList) || targetList.length === 0) {
      errors.push(`${label}.${name}.targets must be a non-empty array`);
      continue;
    }
    const seen = new Set();
    for (const target of targetList) {
      if (typeof target !== "string" || target.trim() === "") {
        errors.push(`${label}.${name}.targets contains an invalid target`);
        continue;
      }
      if (seen.has(target)) {
        errors.push(`${label}.${name}.targets contains duplicate target ${target}`);
      }
      seen.add(target);
      if (!targets.has(target)) {
        errors.push(`${label}.${name}.targets references unknown target ${target}`);
      }
    }
    for (const group of value.groups ?? []) {
      if (typeof group?.name !== "string" || group.name.trim() === "") {
        errors.push(`${label}.${name}.groups contains an invalid group name`);
      }
      if (!Array.isArray(group?.targets) || group.targets.length === 0) {
        errors.push(`${label}.${name}.groups.${group?.name ?? "unknown"}.targets must be non-empty`);
        continue;
      }
      for (const target of group.targets) {
        if (!targets.has(target)) {
          errors.push(`${label}.${name}.groups.${group.name} references unknown target ${target}`);
        }
      }
    }
  }
}

function validateHarnessTiers(errors, harnessChecks, tiers) {
  if (!tiers || typeof tiers !== "object" || Array.isArray(tiers)) {
    errors.push("harness_tiers must be an object");
    return;
  }
  for (const [name, tier] of Object.entries(tiers)) {
    const checks = tier?.checks;
    if (!Array.isArray(checks) || checks.length === 0) {
      errors.push(`harness_tiers.${name}.checks must be a non-empty array`);
      continue;
    }
    const seen = new Set();
    for (const check of checks) {
      if (typeof check !== "string" || check.trim() === "") {
        errors.push(`harness_tiers.${name}.checks contains an invalid check`);
        continue;
      }
      if (seen.has(check)) {
        errors.push(`harness_tiers.${name}.checks contains duplicate check ${check}`);
      }
      seen.add(check);
      if (!harnessChecks.has(check)) {
        errors.push(`harness_tiers.${name}.checks references unknown harness check ${check}`);
      }
    }
  }
}

function validateHelpTiers(errors, targets, tiers) {
  if (!Array.isArray(tiers) || tiers.length === 0) {
    errors.push("help_tiers[] must be a non-empty array");
    return;
  }

  const tierNames = new Set();
  const placements = new Map();
  let dailyTier;

  for (const [tierIndex, tier] of tiers.entries()) {
    const tierLabel = `help_tiers[${tierIndex + 1}]`;
    const tierName = tier?.name;
    if (typeof tierName !== "string" || tierName.trim() === "") {
      errors.push(`${tierLabel}.name must be a non-empty string`);
    } else {
      if (tierNames.has(tierName)) {
        errors.push(`help_tiers contains duplicate tier ${tierName}`);
      }
      tierNames.add(tierName);
      if (tierName === defaultHelpTierName) {
        dailyTier = tier;
      }
    }

    if (!Array.isArray(tier?.entries) || tier.entries.length === 0) {
      errors.push(`${tierLabel}.entries must be a non-empty array`);
      continue;
    }

    const tierTargets = new Set();
    for (const [entryIndex, helpEntry] of tier.entries.entries()) {
      const label = `${tierLabel}.entries[${entryIndex + 1}]`;
      if (typeof helpEntry?.target !== "string" || helpEntry.target.trim() === "") {
        errors.push(`${label}.target must be a non-empty string`);
        continue;
      }
      if (tierTargets.has(helpEntry.target)) {
        errors.push(`${label} contains duplicate target ${helpEntry.target}`);
      }
      tierTargets.add(helpEntry.target);

      const target = targets.get(helpEntry.target);
      if (!target) {
        errors.push(`${label} references unknown target ${helpEntry.target}`);
        continue;
      }
      if (target.classification !== "public") {
        errors.push(`${helpEntry.target} appears in help tier ${tierName ?? "unknown"} but is not classified public`);
      }
      const targetPlacements = placements.get(helpEntry.target) ?? [];
      targetPlacements.push(tierName ?? "unknown");
      placements.set(helpEntry.target, targetPlacements);

      if (typeof helpEntry.description !== "string" || helpEntry.description.trim() === "") {
        errors.push(`${label}.description must be a non-empty string`);
      }
    }
  }

  if (!dailyTier) {
    errors.push(`help_tiers must include a ${defaultHelpTierName} tier`);
  } else if (dailyTier.entries.length > defaultHelpMaxEntries) {
    errors.push(`help_tiers.${defaultHelpTierName} entries must not exceed ${defaultHelpMaxEntries} default help entries`);
  }

  for (const target of targets.values()) {
    if (target.classification !== "public") {
      continue;
    }
    const targetPlacements = placements.get(target.name) ?? [];
    if (targetPlacements.length === 0) {
      errors.push(`public target ${target.name} is missing help tier placement`);
    } else if (targetPlacements.length > 1) {
      errors.push(`public target ${target.name} appears in multiple help tiers: ${targetPlacements.join(",")}`);
    }
  }
}

export function renderTaskSurfaceMake(manifest) {
  const lines = [
    "# Code generated by scripts/render-task-surface-make.mjs; DO NOT EDIT.",
    "",
  ];
  lines.push(`.PHONY: ${targetEntries(manifest).map((entry) => entry.name).join(" ")}`);
  lines.push("");
  lines.push("TASK_SURFACE_HELP_LINES := \\");
  for (const line of helpLines(manifest)) {
    lines.push(`\t'${escapeMakeSingleQuoted(line)}' \\`);
  }
  lines.push("\t''");
  lines.push("");
  lines.push("TASK_SURFACE_HELP_ALL_LINES := \\");
  for (const line of helpAllLines(manifest)) {
    lines.push(`\t'${escapeMakeSingleQuoted(line)}' \\`);
  }
  lines.push("\t''");
  lines.push("");
  return `${lines.join("\n")}\n`;
}

export function helpLines(manifest) {
  const lines = ["Cartulary daily task surface", ""];
  appendHelpTierLines(lines, helpTiers(manifest).find((tier) => tier.name === defaultHelpTierName));
  lines.push("");
  lines.push("For all public targets, run: make help-all");
  lines.push("For private/check internals, run: make task-surface-report TASK_SURFACE_REPORT_ARGS=--all");
  return trimTrailingBlank(lines);
}

export function helpAllLines(manifest) {
  const lines = ["Cartulary public task surface", ""];
  for (const tier of helpTiers(manifest)) {
    appendHelpTierLines(lines, tier);
    lines.push("");
  }
  return trimTrailingBlank(lines);
}

function appendHelpTierLines(lines, tier) {
  if (!tier) {
    return;
  }
  lines.push(`${tier.name}:`);
  for (const entry of tier.entries) {
    lines.push(`  make ${entry.target.padEnd(30)} ${entry.description}`);
  }
}

function trimTrailingBlank(lines) {
  if (lines[lines.length - 1] === "") {
    lines.pop();
  }
  return lines;
}

function escapeMakeSingleQuoted(value) {
  return value.replaceAll("'", "'\"'\"'");
}
