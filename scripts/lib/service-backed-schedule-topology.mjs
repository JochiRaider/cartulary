#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import { loadBrowserBatchStages } from "./browser-batch-manifest.mjs";
import {
  browserStageResource,
  normalizeResourceClaims,
  normalizeResourceLimits,
} from "./scheduler-resources.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..");
const profileSchemaID = "cartulary.service_backed_schedule_profiles.v1";
const scheduleSchemaID = "cartulary.service_backed_schedule.v8";

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

function loadProfile(file) {
  const profile = readJSON(file);
  if (profile.schema_id !== profileSchemaID) {
    throw new Error(`${file} must declare schema_id ${profileSchemaID}`);
  }
  return profile;
}

function loadScheduleManifest(file) {
  const manifest = readJSON(file);
  if (manifest.schema_id !== scheduleSchemaID) {
    throw new Error(`${file} must declare schema_id ${scheduleSchemaID}`);
  }
  return manifest;
}

function assertSameList(actual, expected, label) {
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(`${label} got ${JSON.stringify(actual)} want ${JSON.stringify(expected)}`);
  }
}

function validateResourceShape(schedule) {
  const label = `service-backed schedule ${schedule.target}`;
  const normalized = normalizeResourceLimits(schedule.resource_limits, label, {
    scheduler: "service_backed",
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
  profilePath,
  browserBatchManifestPath,
}) {
  const scheduleManifest = loadScheduleManifest(scheduleManifestPath);
  const profile = loadProfile(profilePath);
  const browserStages = loadBrowserBatchStages(resolveRepoPath(browserBatchManifestPath));
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
    const priorBrowserTargets = [];
    const expectedBrowserTargets = [];

    for (const stageProfile of requireArray(scheduleProfile.browser_stages ?? [], `${scheduleTarget}.browser_stages`)) {
      const stageName = requireString(stageProfile.name, `${scheduleTarget}.browser_stages[].name`);
      const stage = browserStages.get(stageName);
      if (!stage) {
        throw new Error(`${scheduleTarget} references unknown browser batch stage ${stageName}`);
      }
      expectedBrowserTargets.push(stage.target);
      const source = browserSources.find((entry) => entry.browser_stage === stageName);
      if (!source) {
        throw new Error(`${scheduleTarget} must include browser stage ${stageName} target ${stage.target}`);
      }
      validateBrowserSource(schedule, source, stage, resourceLimits);
      const expectedNeeds = [];
      if (stageProfile.needs?.include_backend_targets === true) {
        expectedNeeds.push(...backendTargets);
      }
      if (stageProfile.needs?.include_prior_browser_stages === true) {
        expectedNeeds.push(...priorBrowserTargets);
      }
      assertSameList(source.needs ?? [], expectedNeeds, `${scheduleTarget} ${source.target} needs`);
      priorBrowserTargets.push(stage.target);
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
    "usage: service-backed-schedule-topology.mjs validate <schedule-manifest> <profile> <browser-batch-manifest>",
  );
}

function main(argv) {
  const [command, scheduleManifestPath, profilePath, browserBatchManifestPath] = argv;
  if (command !== "validate" || !scheduleManifestPath || !profilePath || !browserBatchManifestPath) {
    usage();
  }
  validateServiceBackedScheduleTopology({
    scheduleManifestPath,
    profilePath,
    browserBatchManifestPath,
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
