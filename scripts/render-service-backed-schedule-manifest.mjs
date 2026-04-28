#!/usr/bin/env node
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadBrowserBatchStages } from "./lib/browser-batch-manifest.mjs";
import { collectTargetNames, findTargetDescriptor } from "./lib/target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const profileSchemaID = "cartulary.service_backed_schedule_profiles.v1";
const scheduleSchemaID = "cartulary.service_backed_schedule.v7";
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

function cloneObject(value) {
  return JSON.parse(JSON.stringify(value));
}

function laneResourceForStage(stageName) {
  return `browser_stage_${stageName.replaceAll("-", "_")}`;
}

function orderedServiceBackedBackendTargets(profile) {
  const preferred = requireArray(profile.backend_target_order ?? [], "backend_target_order").map(
    (target, index) => requireString(target, `backend_target_order[${index}]`),
  );
  const preferredSet = new Set(preferred);
  const discovered = collectTargetNames(repoRoot).filter((target) => {
    const descriptor = findTargetDescriptor(target, repoRoot);
    return descriptor?.serviceBacked === true && !preferredSet.has(target);
  });
  return [...preferred, ...discovered].filter((target) => {
    const descriptor = findTargetDescriptor(target, repoRoot);
    return descriptor?.serviceBacked === true;
  });
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

function browserSource(profile, stageProfile, stage, backendTargets, priorBrowserTargets) {
  const stageName = requireString(stageProfile.name, "browser_stages[].name");
  const laneResource = laneResourceForStage(stageName);
  const claims = {
    ...cloneObject(profile.defaults.browser_make_target_resource_claims),
    [laneResource]: 1,
  };
  const needs = [];
  if (stageProfile.needs?.include_backend_targets === true) {
    needs.push(...backendTargets);
  }
  if (stageProfile.needs?.include_prior_browser_stages === true) {
    needs.push(...priorBrowserTargets);
  }
  return {
    type: "make_target",
    class: "browser",
    target: stage.target,
    browser_stage: stageName,
    ...(needs.length > 0 ? { needs } : {}),
    weight: stageProfile.weight,
    resource_claims: claims,
  };
}

function renderSchedule(profile, scheduleProfile, browserStages) {
  const target = requireString(scheduleProfile.target, "schedules[].target");
  const resourceLimits = requireObject(scheduleProfile.resource_limits, `${target}.resource_limits`);
  const sources = [];
  if (scheduleProfile.include_service_backed_backend_targets === true) {
    for (const backendTarget of orderedServiceBackedBackendTargets(profile)) {
      sources.push(backendSource(profile, backendTarget));
    }
  }
  const backendTargets = sources
    .filter((source) => source.class === "backend")
    .map((source) => source.target);
  const priorBrowserTargets = [];
  for (const stageProfile of requireArray(scheduleProfile.browser_stages ?? [], `${target}.browser_stages`)) {
    const stageName = requireString(stageProfile.name, `${target}.browser_stages[].name`);
    const stage = browserStages.get(stageName);
    if (!stage) {
      throw new Error(`${target} references unknown browser batch stage ${stageName}`);
    }
    if (!Number.isFinite(stageProfile.weight) || stageProfile.weight < 0) {
      throw new Error(`${target} browser stage ${stageName} must declare a non-negative weight`);
    }
    const source = browserSource(profile, stageProfile, stage, backendTargets, priorBrowserTargets);
    sources.push(source);
    priorBrowserTargets.push(source.target);
  }
  return {
    target,
    resource_limits: cloneObject(resourceLimits),
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
