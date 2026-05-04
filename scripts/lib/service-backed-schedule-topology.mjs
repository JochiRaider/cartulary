#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import { normalizeBrowserBatchStages } from "./browser-batch-manifest.mjs";
import {
  defaultExecutionTopologyManifestPath,
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderServiceBackedScheduleProfile,
} from "./execution-topology.mjs";
import {
  compareExecutionDependencies,
  executionDependencyInfo,
} from "./execution-dependencies.mjs";
import {
  browserStageResource,
  normalizeResourceClaims,
  normalizeResourceLimits,
} from "./scheduler-resources.mjs";
import { validateServiceBackedScheduleManifestShape } from "./service-backed-schedule-manifest.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..");

function resolveRepoPath(file) {
  return path.isAbsolute(file) ? file : path.join(repoRoot, file);
}

function readJSON(file) {
  return JSON.parse(readFileSync(resolveRepoPath(file), "utf8"));
}

function requireArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value;
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function loadScheduleManifest(file) {
  const manifest = readJSON(file);
  return validateServiceBackedScheduleManifestShape(manifest, file);
}

function assertSameList(actual, expected, label) {
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(`${label} got ${JSON.stringify(actual)} want ${JSON.stringify(expected)}`);
  }
}

function requireStringArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  const seen = new Set();
  return value.map((item, index) => {
    const normalized = requireString(item, `${label}[${index}]`);
    if (seen.has(normalized)) {
      throw new Error(`${label} contains duplicate ${normalized}`);
    }
    seen.add(normalized);
    return normalized;
  });
}

function browserSelector(scheduleProfile, scheduleTarget) {
  const selector = scheduleProfile.selectors?.browser;
  if (selector === undefined) {
    return null;
  }
  if (!selector || typeof selector !== "object" || Array.isArray(selector)) {
    throw new Error(`${scheduleTarget}.selectors.browser must be an object when present`);
  }
  const scheduleTags = requireStringArray(selector.schedule_tags, `${scheduleTarget}.selectors.browser.schedule_tags`);
  if (scheduleTags.length === 0) {
    throw new Error(`${scheduleTarget}.selectors.browser.schedule_tags must not be empty`);
  }
  return { scheduleTags };
}

function stageHasRequiredTag(stage, requiredTags) {
  const tags = new Set(stage.scheduleTags ?? []);
  return requiredTags.every((tag) => tags.has(tag));
}

function stageNonRawExecutionDependencies(stage) {
  return Array.from(
    new Set(
      stage.groups
        .filter((group) => group.coverage !== "raw")
        .map((group) => group.executionDependency)
        .filter((dependency) => dependency !== ""),
    ),
  );
}

function validateStageHasServiceBackedEvidence(stage, scheduleTarget) {
  const dependencies = stageNonRawExecutionDependencies(stage);
  const rawOnlyVisual =
    stage.target === "browser-e2e-visual" &&
    stage.groups.every((group) => group.coverage === "raw" && group.kind === "visual");
  if (dependencies.length === 0 && !rawOnlyVisual) {
    throw new Error(`${scheduleTarget} browser stage ${stage.name} has no non-raw execution dependencies`);
  }
  for (const dependency of dependencies) {
    const info = executionDependencyInfo(dependency);
    if (!info || info.category !== "browser" || info.service_backed !== true) {
      throw new Error(
        `${scheduleTarget} browser stage ${stage.name} dependency ${dependency} is not service-backed browser evidence`,
      );
    }
  }
}

function selectedBrowserStages(scheduleProfile, scheduleTarget, browserStages) {
  const selector = browserSelector(scheduleProfile, scheduleTarget);
  if (!selector) {
    return [];
  }
  const stages = Array.from(browserStages.values()).filter((stage) =>
    stageHasRequiredTag(stage, selector.scheduleTags),
  ).sort((left, right) => {
    const leftDependency = stageNonRawExecutionDependencies(left).sort(compareExecutionDependencies)[0] ?? "";
    const rightDependency = stageNonRawExecutionDependencies(right).sort(compareExecutionDependencies)[0] ?? "";
    return (
      compareExecutionDependencies(leftDependency, rightDependency) ||
      left.name.localeCompare(right.name)
    );
  });
  for (const stage of stages) {
    validateStageHasServiceBackedEvidence(stage, scheduleTarget);
  }
  return stages;
}

function expectedNeedsForStage(stage, selectedTargets, scheduleTarget) {
  const needs = stage.schedulerNeeds ?? [];
  for (const need of needs) {
    if (need === stage.target) {
      throw new Error(`${scheduleTarget} browser stage ${stage.name} must not depend on itself`);
    }
    if (!selectedTargets.has(need)) {
      throw new Error(
        `${scheduleTarget} browser stage ${stage.name} scheduler_needs target ${need} is not selected by the schedule`,
      );
    }
  }
  return needs;
}

function validateResourceShape(schedule) {
  const label = `service-backed schedule ${schedule.target}`;
  const normalized = normalizeResourceLimits(schedule.resource_limits, label, {
    scheduler: "service_backed",
    capacityProfile: schedule.capacity_profile ?? null,
    allowAuto: true,
  });
  const resourceLimits = normalized.limits;
  for (const resource of ["go_cpu", "go_io"]) {
    if (!resourceLimits.has(resource)) {
      throw new Error(`${label} must declare ${resource} resource limit`);
    }
  }

  for (const [index, source] of requireArray(schedule.work_unit_sources, `${label}.work_unit_sources`).entries()) {
    const sourceLabel = `${label} work_unit_sources ${index + 1} ${source?.target ?? ""}`.trim();
    if (!source || typeof source !== "object" || Array.isArray(source)) {
      throw new Error(`${sourceLabel} must be an object`);
    }
    const claims = normalizeResourceClaims(
      source.resource_claims,
      sourceLabel,
      resourceLimits,
      { scheduler: "service_backed" },
    );
    if (source.type === "go_shards" && (claims.has("go_cpu") || claims.has("go_io"))) {
      throw new Error(`${sourceLabel} go shard source must leave go_cpu/go_io to per-shard scheduler profiles`);
    }
  }
  return resourceLimits;
}

function validateBrowserSource(schedule, source, stage, resourceLimits) {
  const label = `service-backed schedule ${schedule.target} ${source.target}`;
  const stageResource = browserStageResource(stage.name);
  const claims = normalizeResourceClaims(source.resource_claims, label, resourceLimits, {
    scheduler: "service_backed",
  });
  if (source.type !== "make_target") {
    throw new Error(`${label} must use make_target for browser work`);
  }
  if (source.browser_stage !== stage.name) {
    throw new Error(`${label} must declare browser_stage ${stage.name}`);
  }
  if (source.target !== stage.target) {
    throw new Error(`${label} must match browser stage target ${stage.target}`);
  }
  if (!resourceLimits.has("browser_stack")) {
    throw new Error(`${schedule.target} must declare browser_stack resource limit for browser work`);
  }
  if (!resourceLimits.has(stageResource)) {
    throw new Error(`${schedule.target} must declare ${stageResource} resource limit for ${source.target}`);
  }
  if (!claims.has("browser_stack")) {
    throw new Error(`${label} must claim browser_stack`);
  }
  if (!claims.has(stageResource)) {
    throw new Error(`${label} must claim ${stageResource}`);
  }
}

export function validateServiceBackedScheduleTopology({
  scheduleManifestPath,
  topologyPath = defaultExecutionTopologyManifestPath,
}) {
  const scheduleManifest = loadScheduleManifest(scheduleManifestPath);
  const topology = loadExecutionTopology({ manifestPath: topologyPath });
  const profile = renderServiceBackedScheduleProfile(topology);
  const browserStages = normalizeBrowserBatchStages(renderBrowserBatchManifest(topology));
  const schedulesByTarget = new Map();
  for (const schedule of requireArray(scheduleManifest.schedules, "schedules")) {
    const target = requireString(schedule?.target, "schedules[].target");
    if (schedulesByTarget.has(target)) {
      throw new Error(`service-backed schedule manifest declares duplicate schedule ${target}`);
    }
    schedulesByTarget.set(target, schedule);
  }

  for (const scheduleProfile of requireArray(profile.schedules, "profile.schedules")) {
    const scheduleTarget = requireString(scheduleProfile.target, "profile.schedules[].target");
    const schedule = schedulesByTarget.get(scheduleTarget);
    if (!schedule) {
      throw new Error(`service-backed schedule manifest is missing profile schedule ${scheduleTarget}`);
    }
    const resourceLimits = validateResourceShape(schedule);
    const backendTargets = requireArray(schedule.work_unit_sources, `${scheduleTarget}.work_unit_sources`)
      .filter((source) => source.class === "backend")
      .map((source) => source.target);
    const browserSources = schedule.work_unit_sources.filter((source) => source.class === "browser");
    const expectedBrowserTargets = [];
    const selectedStages = selectedBrowserStages(scheduleProfile, scheduleTarget, browserStages);
    const selectedTargets = new Set([
      ...backendTargets,
      ...selectedStages.map((stage) => stage.target),
    ]);

    for (const stage of selectedStages) {
      expectedBrowserTargets.push(stage.target);
      const source = browserSources.find((entry) => entry.browser_stage === stage.name);
      if (!source) {
        throw new Error(`${scheduleTarget} must include browser stage ${stage.name} target ${stage.target}`);
      }
      validateBrowserSource(schedule, source, stage, resourceLimits);
      const expectedNeeds = expectedNeedsForStage(stage, selectedTargets, scheduleTarget);
      assertSameList(source.needs ?? [], expectedNeeds, `${scheduleTarget} ${source.target} needs`);
    }

    assertSameList(
      browserSources.map((source) => source.target),
      expectedBrowserTargets,
      `${scheduleTarget} browser targets`,
    );
  }
}

function usage() {
  throw new Error(
    "usage: service-backed-schedule-topology.mjs validate <schedule-manifest> [topology-manifest]",
  );
}

function main(argv) {
  const [command, scheduleManifestPath, topologyPath = defaultExecutionTopologyManifestPath] = argv;
  if (command !== "validate" || !scheduleManifestPath) {
    usage();
  }
  validateServiceBackedScheduleTopology({
    scheduleManifestPath,
    topologyPath,
  });
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exit(1);
  }
}
