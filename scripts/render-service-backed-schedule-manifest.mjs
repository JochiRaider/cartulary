#!/usr/bin/env node
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadBrowserBatchStages } from "./lib/browser-batch-manifest.mjs";
import {
  compareExecutionDependencies,
  executionDependencyInfo,
} from "./lib/execution-dependencies.mjs";
import {
  collectEntries,
  loadManifest,
  phaseManifestNames,
} from "./lib/phase-manifest.mjs";
import { browserStageResource } from "./lib/scheduler-resources.mjs";
import { collectTargetPlanRows, findTargetDescriptor } from "./lib/target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const profileSchemaID = "cartulary.service_backed_schedule_profiles.v2";
const scheduleSchemaID = "cartulary.service_backed_schedule.v8";
const defaultProfilePath = path.join(repoRoot, "tools", "service_backed_schedule_profiles.json");
const defaultOutputPath = path.join(repoRoot, "tools", "service_backed_schedule_manifest.json");
const defaultBrowserBatchManifestPath = path.join(repoRoot, "tools", "browser_e2e_batch_manifest.json");

function usage() {
  throw new Error(
    "usage: render-service-backed-schedule-manifest.mjs [--check] [--profile <path>] [--output <path>] [--browser-batch-manifest <path>]",
  );
}

function parseArgs(argv) {
  const options = {
    check: false,
    profile: defaultProfilePath,
    output: defaultOutputPath,
    browserBatchManifest: defaultBrowserBatchManifestPath,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--check") {
      options.check = true;
      continue;
    }
    if (arg === "--profile") {
      options.profile = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--output") {
      options.output = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--browser-batch-manifest") {
      options.browserBatchManifest = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.profile || !options.output || !options.browserBatchManifest) {
    usage();
  }
  return options;
}

function resolvePath(file) {
  return path.isAbsolute(file) ? file : path.join(repoRoot, file);
}

function readJSON(file) {
  return JSON.parse(readFileSync(resolvePath(file), "utf8"));
}

function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
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

function requireBoolean(value, label) {
  if (typeof value !== "boolean") {
    throw new Error(`${label} must be a boolean`);
  }
  return value;
}

function requireStringArray(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  const seen = new Set();
  const result = [];
  for (const [index, item] of value.entries()) {
    const normalized = requireString(item, `${label}[${index}]`);
    if (seen.has(normalized)) {
      throw new Error(`${label} must not contain duplicate ${normalized}`);
    }
    seen.add(normalized);
    result.push(normalized);
  }
  return result;
}

function cloneObject(value) {
  return JSON.parse(JSON.stringify(value));
}

function minExecutionDependency(target) {
  const descriptor = findTargetDescriptor(target, repoRoot);
  const dependencies = [
    ...(descriptor?.executionDependencies ?? []),
    ...(descriptor?.supportTargets ?? []),
  ].filter((dependency) => dependency !== "");
  if (dependencies.length === 0) {
    return "";
  }
  return dependencies.sort(compareExecutionDependencies)[0];
}

function compareBackendTargets(left, right) {
  const leftDependency = minExecutionDependency(left);
  const rightDependency = minExecutionDependency(right);
  return (
    compareExecutionDependencies(leftDependency, rightDependency) ||
    String(left).localeCompare(String(right))
  );
}

function backendSelector(scheduleProfile) {
  const selector = scheduleProfile.selectors?.backend ?? {};
  if (!selector || typeof selector !== "object" || Array.isArray(selector)) {
    throw new Error(`${scheduleProfile.target}.selectors.backend must be an object when present`);
  }
  return {
    serviceBacked:
      selector.service_backed === undefined
        ? true
        : requireBoolean(selector.service_backed, `${scheduleProfile.target}.selectors.backend.service_backed`),
    checkServiceBackedSafe:
      selector.check_service_backed_safe === undefined
        ? true
        : requireBoolean(
            selector.check_service_backed_safe,
            `${scheduleProfile.target}.selectors.backend.check_service_backed_safe`,
          ),
  };
}

function browserSelector(scheduleProfile) {
  const selector = scheduleProfile.selectors?.browser;
  if (selector === undefined) {
    return null;
  }
  if (!selector || typeof selector !== "object" || Array.isArray(selector)) {
    throw new Error(`${scheduleProfile.target}.selectors.browser must be an object when present`);
  }
  const tags = requireStringArray(selector.schedule_tags, `${scheduleProfile.target}.selectors.browser.schedule_tags`);
  if (tags.length === 0) {
    throw new Error(`${scheduleProfile.target}.selectors.browser.schedule_tags must not be empty`);
  }
  return { scheduleTags: tags };
}

function orderedServiceBackedBackendTargets(scheduleProfile) {
  const selector = backendSelector(scheduleProfile);
  const targetsWithRows = new Set(
    collectTargetPlanRows(repoRoot)
      .filter((row) => {
        if (row.runner_family !== "go_test") {
          return false;
        }
        if (selector.serviceBacked && row.service_backed !== true) {
          return false;
        }
        if (selector.checkServiceBackedSafe && row.check_service_backed_safe !== true) {
          return false;
        }
        return true;
      })
      .map((row) => row.target),
  );
  return Array.from(targetsWithRows)
    .filter((target) => {
      const descriptor = findTargetDescriptor(target, repoRoot);
      return (
        descriptor?.serviceBacked === selector.serviceBacked &&
        (!selector.checkServiceBackedSafe || descriptor?.checkServiceBackedSafe === true)
      );
    })
    .sort(compareBackendTargets);
}

function backendSource(profile, target) {
  const descriptor = findTargetDescriptor(target, repoRoot);
  if (!descriptor) {
    throw new Error(`unknown backend target ${target}`);
  }
  if (!descriptor.serviceBacked) {
    throw new Error(`backend target ${target} is not service-backed`);
  }
  if (descriptor.sharding === "go_shards") {
    return {
      type: "go_shards",
      class: "backend",
      target,
      resource_claims: cloneObject(profile.defaults.go_shards_resource_claims),
    };
  }
  const weights = requireObject(
    profile.defaults.backend_make_target_weights,
    "defaults.backend_make_target_weights",
  );
  const claims = requireObject(
    profile.defaults.backend_make_target_resource_claims,
    "defaults.backend_make_target_resource_claims",
  );
  if (!Number.isFinite(weights[target])) {
    throw new Error(`defaults.backend_make_target_weights must declare ${target}`);
  }
  if (!claims[target]) {
    throw new Error(`defaults.backend_make_target_resource_claims must declare ${target}`);
  }
  return {
    type: "make_target",
    class: "backend",
    target,
    weight: weights[target],
    resource_claims: cloneObject(claims[target]),
  };
}

function includeBackendDependencies(policy) {
  return policy === "after_backend" || policy === "after_backend_and_prior_browser";
}

function includePriorBrowserDependencies(policy) {
  return policy === "after_prior_browser" || policy === "after_backend_and_prior_browser";
}

function browserSource(profile, stage, backendTargets, priorBrowserTargets, weight) {
  const stageName = stage.name;
  const laneResource = browserStageResource(stageName);
  const claims = {
    ...cloneObject(profile.defaults.browser_make_target_resource_claims),
    [laneResource]: 1,
  };
  const needs = [];
  if (includeBackendDependencies(stage.schedulerDependencyPolicy)) {
    needs.push(...backendTargets);
  }
  if (includePriorBrowserDependencies(stage.schedulerDependencyPolicy)) {
    needs.push(...priorBrowserTargets);
  }
  return {
    type: "make_target",
    class: "browser",
    target: stage.target,
    browser_stage: stageName,
    ...(needs.length > 0 ? { needs } : {}),
    weight,
    resource_claims: claims,
  };
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
  if (dependencies.length === 0) {
    throw new Error(`${scheduleTarget} browser stage ${stage.name} has no non-raw execution dependencies`);
  }
  for (const group of stage.groups.filter((candidate) => candidate.coverage !== "raw")) {
    const dependency = group.executionDependency;
    const info = executionDependencyInfo(dependency);
    if (!info || info.category !== "browser" || info.service_backed !== true) {
      throw new Error(
        `${scheduleTarget} browser stage ${stage.name} dependency ${dependency} is not service-backed browser evidence`,
      );
    }
    if (!hasPlaywrightRows(group.coverage, dependency)) {
      throw new Error(
        `${scheduleTarget} browser stage ${stage.name} has no phase-map Playwright rows for ${group.coverage} ${dependency}`,
      );
    }
  }
}

function hasPlaywrightRows(coverage, executionDependency) {
  for (const phase of phaseManifestNames(repoRoot)) {
    const { manifest } = loadManifest(repoRoot, phase);
    if (
      collectEntries(manifest).some(
        (entry) =>
          entry.section === "e2e" &&
          entry.runner === "playwright" &&
          entry.coverage === coverage &&
          entry.execution_dependency === executionDependency,
      )
    ) {
      return true;
    }
  }
  return false;
}

function stageWeight(profile, stage, scheduleTarget) {
  const weights = profile.defaults.browser_stage_weights ?? {};
  if (!weights || typeof weights !== "object" || Array.isArray(weights)) {
    throw new Error("defaults.browser_stage_weights must be an object when present");
  }
  if (!Number.isFinite(weights[stage.name]) || weights[stage.name] < 0) {
    throw new Error(`defaults.browser_stage_weights must declare non-negative weight for ${scheduleTarget} ${stage.name}`);
  }
  return weights[stage.name];
}

function selectedBrowserStages(scheduleProfile, browserStages) {
  const selector = browserSelector(scheduleProfile);
  if (!selector) {
    return [];
  }
  const stages = Array.from(browserStages.values())
    .filter((stage) => stageHasRequiredTag(stage, selector.scheduleTags))
    .sort((left, right) => {
      const leftDependency = stageNonRawExecutionDependencies(left).sort(compareExecutionDependencies)[0] ?? "";
      const rightDependency = stageNonRawExecutionDependencies(right).sort(compareExecutionDependencies)[0] ?? "";
      return (
        compareExecutionDependencies(leftDependency, rightDependency) ||
        left.name.localeCompare(right.name)
      );
    });
  for (const stage of stages) {
    validateStageHasServiceBackedEvidence(stage, scheduleProfile.target);
  }
  return stages;
}

function renderSchedule(profile, scheduleProfile, browserStages) {
  const target = requireString(scheduleProfile.target, "schedules[].target");
  const resourceLimits = cloneObject(requireObject(scheduleProfile.resource_limits, `${target}.resource_limits`));
  const sources = [];
  for (const backendTarget of orderedServiceBackedBackendTargets(scheduleProfile)) {
    sources.push(backendSource(profile, backendTarget));
  }
  const backendTargets = sources
    .filter((source) => source.class === "backend")
    .map((source) => source.target);
  const priorBrowserTargets = [];
  for (const stage of selectedBrowserStages(scheduleProfile, browserStages)) {
    resourceLimits[browserStageResource(stage.name)] = 1;
    const source = browserSource(
      profile,
      stage,
      backendTargets,
      priorBrowserTargets,
      stageWeight(profile, stage, target),
    );
    sources.push(source);
    priorBrowserTargets.push(source.target);
  }
  return {
    target,
    resource_limits: resourceLimits,
    work_unit_sources: sources,
  };
}

function renderManifest(options) {
  const profile = readJSON(options.profile);
  if (profile.schema_id !== profileSchemaID) {
    throw new Error(`${options.profile} must declare schema_id ${profileSchemaID}`);
  }
  requireObject(profile.defaults, "defaults");
  const browserStages = loadBrowserBatchStages(resolvePath(options.browserBatchManifest));
  return {
    schema_id: scheduleSchemaID,
    generated: {
      generator: "scripts/render-service-backed-schedule-manifest.mjs",
      profile: path.relative(repoRoot, resolvePath(options.profile)),
      browser_batch_manifest: path.relative(repoRoot, resolvePath(options.browserBatchManifest)),
    },
    schedules: requireArray(profile.schedules, "schedules").map((schedule) =>
      renderSchedule(profile, schedule, browserStages),
    ),
  };
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const rendered = `${JSON.stringify(renderManifest(options), null, 2)}\n`;
  const outputPath = resolvePath(options.output);
  if (options.check) {
    const existing = readFileSync(outputPath, "utf8");
    if (existing !== rendered) {
      throw new Error(`${path.relative(repoRoot, outputPath)} is stale; run make phase-schedules`);
    }
    return;
  }
  writeFileSync(outputPath, rendered);
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`service-backed schedule render failed: ${message}`);
  process.exit(1);
}
