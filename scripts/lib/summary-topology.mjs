import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadBrowserBatchStages, normalizeBrowserBatchStages } from "./browser-batch-manifest.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");
export const defaultTaskSurfaceManifestPath = path.join(repoRoot, "tools", "task_surface_manifest.json");
export const defaultServiceBackedScheduleManifestPath = path.join(
  repoRoot,
  "tools",
  "service_backed_schedule_manifest.json",
);
export const defaultBrowserBatchManifestPath = path.join(repoRoot, "tools", "browser_e2e_batch_manifest.json");
export const serviceBackedScheduleSchemaID = "cartulary.service_backed_schedule.v8";

function resolveRepoPath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function readJSON(file) {
  return JSON.parse(readFileSync(resolveRepoPath(file), "utf8"));
}

function configuredPath(value, fallback) {
  return typeof value === "string" && value.trim() !== "" ? value : fallback;
}

function requireTargetList(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  const seen = new Set();
  return value.map((entry, index) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label}[${index + 1}] must be a non-empty string`);
    }
    const target = entry.trim();
    if (seen.has(target)) {
      throw new Error(`${label} contains duplicate target ${target}`);
    }
    seen.add(target);
    return target;
  });
}

function loadServiceBackedScheduleManifest(file = defaultServiceBackedScheduleManifestPath) {
  const manifest = readJSON(file);
  if (manifest.schema_id !== serviceBackedScheduleSchemaID) {
    throw new Error(`${resolveRepoPath(file)} must declare schema_id ${serviceBackedScheduleSchemaID}`);
  }
  if (!Array.isArray(manifest.schedules)) {
    throw new Error(`${resolveRepoPath(file)} must declare schedules[]`);
  }
  return manifest;
}

export function loadSummaryTopologyContext(options = {}) {
  const taskSurfaceManifest =
    options.taskSurfaceManifest ??
    (configuredPath(options.taskSurfaceManifestPath, "") ? readJSON(options.taskSurfaceManifestPath) : null);
  const serviceBackedScheduleManifest =
    options.serviceBackedScheduleManifest ??
    loadServiceBackedScheduleManifest(
      configuredPath(options.serviceBackedScheduleManifestPath, defaultServiceBackedScheduleManifestPath),
    );
  const browserBatchManifestPath = resolveRepoPath(
    configuredPath(options.browserBatchManifestPath, defaultBrowserBatchManifestPath),
  );
  const browserStages =
    options.browserStages ??
    (options.browserBatchManifest
      ? normalizeBrowserBatchStages(options.browserBatchManifest)
      : loadBrowserBatchStages(browserBatchManifestPath));
  return {
    taskSurfaceManifest,
    serviceBackedScheduleManifest,
    browserStages,
  };
}

function findServiceBackedSchedule(context, target) {
  const matches = (context.serviceBackedScheduleManifest?.schedules ?? []).filter(
    (schedule) => schedule?.target === target,
  );
  if (matches.length === 0) {
    return null;
  }
  if (matches.length > 1) {
    throw new Error(`expected at most one service-backed schedule for ${target}, found ${matches.length}`);
  }
  return matches[0];
}

function scheduleSources(schedule, target) {
  if (!Array.isArray(schedule?.work_unit_sources)) {
    throw new Error(`service-backed schedule ${target} must declare work_unit_sources[]`);
  }
  return schedule.work_unit_sources.map((source, index) => {
    if (!source || typeof source !== "object" || Array.isArray(source)) {
      throw new Error(`service-backed schedule ${target} work_unit_sources[${index + 1}] must be an object`);
    }
    if (typeof source.target !== "string" || source.target.trim() === "") {
      throw new Error(`service-backed schedule ${target} work_unit_sources[${index + 1}] must declare target`);
    }
    return {
      target: source.target.trim(),
      class: String(source.class ?? "").trim(),
    };
  });
}

export function serviceBackedScheduleChildren(context, target) {
  const schedule = findServiceBackedSchedule(context, target);
  if (!schedule) {
    return [];
  }
  return scheduleSources(schedule, target).map((source) => source.target);
}

function browserStageByTarget(context, target) {
  for (const stage of context.browserStages.values()) {
    if (stage.target === target) {
      return stage;
    }
  }
  return null;
}

export function browserSummaryChildren(context, target) {
  return [...(browserStageByTarget(context, target)?.summaryChildren ?? [])];
}

function browserLeafTargets(context, targets) {
  const leaves = [];
  for (const target of targets) {
    const children = browserSummaryChildren(context, target);
    leaves.push(...(children.length > 0 ? children : [target]));
  }
  return leaves;
}

function harnessTierChecks(taskSurfaceManifest, tier) {
  const checks = taskSurfaceManifest?.harness_tiers?.[tier]?.checks;
  return checks === undefined ? [] : requireTargetList(checks, `harness_tiers.${tier}.checks`);
}

export function harnessSummaryTargets(context, tier) {
  return harnessTierChecks(context.taskSurfaceManifest, tier);
}

export function harnessSummaryGroups(context, tier) {
  if (tier === "full") {
    return ["fast", "extended", "lifecycle"]
      .map((name) => ({ name, summaryTargets: harnessSummaryTargets(context, name) }))
      .filter((group) => group.summaryTargets.length > 0);
  }
  return [{ name: tier, summaryTargets: harnessSummaryTargets(context, tier) }];
}

function harnessProjectionChildren(context, target) {
  if (target === "check-harness-smoke") {
    return harnessSummaryTargets(context, "fast");
  }
  const match = /^run-harness-smoke-(.+)$/.exec(target);
  if (!match) {
    return [];
  }
  return harnessSummaryTargets(context, match[1]);
}

function summaryEntries(taskSurfaceManifest) {
  return [...(taskSurfaceManifest?.targets ?? []), ...(taskSurfaceManifest?.harness_checks ?? [])];
}

function explicitProjectionChildren(context, target) {
  const entry = summaryEntries(context.taskSurfaceManifest).find((candidate) => candidate?.name === target);
  return entry?.summary_projection?.children ?? [];
}

export function summaryProjectionChildren(context, target) {
  const serviceBackedChildren = serviceBackedScheduleChildren(context, target);
  if (serviceBackedChildren.length > 0) {
    return serviceBackedChildren;
  }
  const browserChildren = browserSummaryChildren(context, target);
  if (browserChildren.length > 0) {
    return browserChildren;
  }
  const harnessChildren = harnessProjectionChildren(context, target);
  if (harnessChildren.length > 0) {
    return harnessChildren;
  }
  return [...explicitProjectionChildren(context, target)];
}

export function resolveSummaryGroups(context, groups = []) {
  if (!Array.isArray(groups)) {
    throw new Error("summary_groups must be an array");
  }
  return groups.map((group, index) => {
    if (!group || typeof group !== "object" || Array.isArray(group)) {
      throw new Error(`summary_groups[${index + 1}] must be an object`);
    }
    if (typeof group.name !== "string" || group.name.trim() === "") {
      throw new Error(`summary_groups[${index + 1}].name must be a non-empty string`);
    }
    if (group.summary_targets !== undefined) {
      return {
        name: group.name.trim(),
        summaryTargets: requireTargetList(group.summary_targets, `summary_groups.${group.name}.summary_targets`),
      };
    }
    const source = group.source;
    if (!source || typeof source !== "object" || Array.isArray(source)) {
      throw new Error(`summary_groups.${group.name}.source must be an object when summary_targets is absent`);
    }
    if (source.type !== "service_backed_schedule") {
      throw new Error(`summary_groups.${group.name}.source.type must be service_backed_schedule`);
    }
    if (typeof source.target !== "string" || source.target.trim() === "") {
      throw new Error(`summary_groups.${group.name}.source.target must be a non-empty string`);
    }
    const target = source.target.trim();
    const sources = scheduleSources(findServiceBackedSchedule(context, target), target);
    if (source.class === "backend") {
      return {
        name: group.name.trim(),
        summaryTargets: sources.filter((entry) => entry.class === "backend").map((entry) => entry.target),
      };
    }
    if (source.class === "browser_leaves") {
      return {
        name: group.name.trim(),
        summaryTargets: browserLeafTargets(
          context,
          sources.filter((entry) => entry.class === "browser").map((entry) => entry.target),
        ),
      };
    }
    throw new Error(`summary_groups.${group.name}.source.class must be backend or browser_leaves`);
  });
}

export function summaryGroupsSpec(groups) {
  return groups.map((group) => `${group.name}=${group.summaryTargets.join(",")}`).join(";");
}

export function collectExplicitSummaryProjectionErrors(taskSurfaceManifest, context) {
  const errors = [];
  const targetNames = new Set(summaryEntries(taskSurfaceManifest).map((entry) => entry?.name).filter(Boolean));
  for (const entry of summaryEntries(taskSurfaceManifest)) {
    if (entry?.summary_projection === undefined) {
      continue;
    }
    const label = `${entry.name}.summary_projection`;
    const children = entry.summary_projection?.children;
    if (!Array.isArray(children)) {
      errors.push(`${label}.children must be an array`);
      continue;
    }
    const derivedChildren = [
      ...serviceBackedScheduleChildren(context, entry.name),
      ...browserSummaryChildren(context, entry.name),
      ...harnessProjectionChildren(context, entry.name),
    ];
    if (derivedChildren.length > 0) {
      errors.push(`${label} duplicates derived summary topology; remove explicit children`);
    }
    for (const child of children) {
      if (!targetNames.has(child)) {
        errors.push(`${entry.name} summary projection references unknown child target ${child}`);
      }
    }
  }
  return errors;
}
