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
export const taskSurfaceSchemaID = "cartulary.task_surface_manifest.v2";

const validClassifications = new Set(["public", "check_internal", "helper_only"]);
const validInclusions = new Set(["test", "check", "ci", "release-check", "helper_only"]);

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
  const entry = targetEntryMap(manifest).get(target);
  const children = entry?.summary_projection?.children ?? [];
  return [...children];
}

export function harnessTierTargets(manifest, name) {
  const tier = manifest.harness_tiers?.[name];
  if (!tier) {
    throw new Error(`unknown harness tier ${name}`);
  }
  return [...tier.targets];
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

  for (const entry of targets.values()) {
    for (const child of entry.summary_projection?.children ?? []) {
      if (!targets.has(child)) {
        errors.push(`${entry.name} summary projection references unknown child target ${child}`);
      }
    }
  }

  if (!Array.isArray(manifest.help_sections) || manifest.help_sections.length === 0) {
    errors.push("help_sections[] must be a non-empty array");
  } else {
    const helped = new Set();
    for (const [sectionIndex, section] of manifest.help_sections.entries()) {
      if (typeof section?.name !== "string" || section.name.trim() === "") {
        errors.push(`help_sections[${sectionIndex + 1}].name must be a non-empty string`);
      }
      if (!Array.isArray(section?.entries)) {
        errors.push(`help_sections[${sectionIndex + 1}].entries must be an array`);
        continue;
      }
      for (const [entryIndex, helpEntry] of section.entries.entries()) {
        const label = `help_sections[${sectionIndex + 1}].entries[${entryIndex + 1}]`;
        if (typeof helpEntry?.target !== "string" || helpEntry.target.trim() === "") {
          errors.push(`${label}.target must be a non-empty string`);
          continue;
        }
        const target = targets.get(helpEntry.target);
        if (!target) {
          errors.push(`${label} references unknown target ${helpEntry.target}`);
          continue;
        }
        helped.add(helpEntry.target);
        if (target.classification !== "public") {
          errors.push(`${helpEntry.target} has help text but is not classified public`);
        }
        if (typeof helpEntry.description !== "string" || helpEntry.description.trim() === "") {
          errors.push(`${label}.description must be a non-empty string`);
        }
      }
    }
    for (const target of targets.values()) {
      if (target.classification === "public" && !helped.has(target.name)) {
        errors.push(`public target ${target.name} is missing help text`);
      }
    }
  }

  validateNamedTargetLists(errors, targets, manifest.summary_profiles, "summary_profiles");
  validateNamedTargetLists(errors, targets, manifest.harness_tiers, "harness_tiers");

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
  for (const [name, tier] of Object.entries(manifest.harness_tiers ?? {})) {
    lines.push(
      `TASK_SURFACE_HARNESS_TIER_${makeIdentifier(name)}_TARGETS := ${tier.targets.join(" ")}`,
    );
  }
  lines.push("");
  return `${lines.join("\n")}\n`;
}

export function helpLines(manifest) {
  const lines = ["Cartulary developer task surface", ""];
  for (const section of manifest.help_sections) {
    lines.push(`${section.name}:`);
    for (const entry of section.entries) {
      lines.push(`  make ${entry.target.padEnd(22)} ${entry.description}`);
    }
    lines.push("");
  }
  if (lines[lines.length - 1] === "") {
    lines.pop();
  }
  return lines;
}

function escapeMakeSingleQuoted(value) {
  return value.replaceAll("'", "'\"'\"'");
}
