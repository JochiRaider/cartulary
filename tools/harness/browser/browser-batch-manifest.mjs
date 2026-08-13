#!/usr/bin/env node
import { readFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

import {
  readJsonObject,
  requireObject,
  requireSchemaID,
  requireString,
  validateObjectArray,
  validateObjectShape,
} from "../contract/json-shape.mjs";

export const browserBatchManifestSchemaID = "cartulary.browser_e2e_batch_manifest.v9";

const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const browserBatchKeys = new Set(["schema_id", "runtime_profiles", "stages"]);
const browserStageKeys = new Set([
  "name",
  "target",
  "schedule_tags",
  "scheduler_dependency_policy",
  "scheduler_needs",
  "summary_children",
  "groups",
]);
const browserGroupKeys = new Set([
  "name",
  "kind",
  "target",
  "project",
  "workers",
  "config",
  "coverage",
  "execution_dependency",
  "dependency_target",
  "reset_before",
  "selected_row_ids",
  "browser_session_group",
  "browser_session_isolation_reason",
  "runtime_profile_id",
  "resource_profile_id",
  "service_dependencies",
  "service_requirement",
  "specs",
]);
const allowedGroupKinds = new Set([
  "webserver-backed",
  "duration_balanced_specs",
  "functional",
  "support",
  "stateful",
  "stateful_partition",
  "measurement",
  "visual",
  "a11y",
]);
const allowedCoverage = new Set(["authoritative", "supplemental", "raw"]);
function manifestValue(fileOrManifest, label) {
  return typeof fileOrManifest === "string"
    ? readJsonObject(fileOrManifest, label)
    : requireObject(fileOrManifest, label);
}

export function validateBrowserBatchManifestShape(fileOrManifest, label = fileOrManifest) {
  const manifest = manifestValue(fileOrManifest, label);
  validateObjectShape(manifest, label, { keys: browserBatchKeys });
  requireSchemaID(manifest, browserBatchManifestSchemaID, label);
  validateObjectArray(
    manifest.runtime_profiles,
    `${label}.runtime_profiles`,
    {
      nonEmpty: true,
      keys: new Set([
        "id",
        "kind",
        "key_ring_manifest_path",
        "service_requirement",
      ]),
    },
    (profile, profileLabel) => {
      requireString(profile.id, `${profileLabel}.id`, { pattern: /^[a-z][a-z0-9_]*$/u });
      requireString(profile.kind, `${profileLabel}.kind`, { pattern: /^[a-z][a-z0-9_]*$/u });
      if (profile.key_ring_manifest_path !== undefined) {
        requireString(profile.key_ring_manifest_path, `${profileLabel}.key_ring_manifest_path`);
      }
      requireString(
        profile.service_requirement,
        `${profileLabel}.service_requirement`,
        { pattern: /^(?:none|test-services)$/u },
      );
    },
  );
  validateObjectArray(
    manifest.stages,
    `${label}.stages`,
    { nonEmpty: true, keys: browserStageKeys },
    (stage, stageLabel) => {
      requireString(stage.name, `${stageLabel}.name`);
      requireString(stage.target, `${stageLabel}.target`, {
        pattern: makeTargetPattern,
      });
      if (stage.scheduler_dependency_policy !== undefined) {
        throw new Error(`${stageLabel}.scheduler_dependency_policy is obsolete; use scheduler_needs[]`);
      }
      validateObjectArray(
        stage.groups,
        `${stageLabel}.groups`,
        { nonEmpty: true, keys: browserGroupKeys },
        (group, groupLabel) => {
          requireString(group.name, `${groupLabel}.name`);
          requireString(group.target, `${groupLabel}.target`, {
            pattern: makeTargetPattern,
          });
        },
      );
    },
  );
  return manifest;
}

export function loadBrowserBatchManifest(manifestPath) {
  return validateBrowserBatchManifestShape(
    JSON.parse(readFileSync(manifestPath, "utf8")),
    manifestPath,
  );
}

export function loadBrowserBatchStages(manifestPath) {
  const manifest = loadBrowserBatchManifest(manifestPath);
  return normalizeBrowserBatchStages(manifest);
}

export function normalizeBrowserBatchStages(manifest) {
  const runtimeProfiles = new Map(
    (manifest.runtime_profiles ?? []).map((profile) => [
      profile.id,
      profile.service_requirement,
    ]),
  );
  if (!runtimeProfiles.has("default")) {
    throw new Error("browser E2E batch manifest must declare the default runtime profile");
  }
  const stages = new Map();
  for (const [index, rawStage] of manifest.stages.entries()) {
    const stage = normalizeStage(rawStage, index + 1, runtimeProfiles);
    if (stages.has(stage.name)) {
      throw new Error(`duplicate browser batch stage ${stage.name}`);
    }
    stages.set(stage.name, stage);
  }
  return stages;
}

export function resolveBrowserBatchStage(manifestPath, stageName) {
  const stages = loadBrowserBatchStages(manifestPath);
  const stage = stages.get(stageName);
  if (!stage) {
    throw new Error(`expected exactly one browser E2E batch stage ${stageName}, found 0`);
  }
  return stage;
}

function normalizeStage(stage, index, runtimeProfiles) {
  const label = `browser E2E batch stage ${index}`;
  if (!stage || typeof stage !== "object" || Array.isArray(stage)) {
    throw new Error(`${label} must be an object`);
  }
  for (const field of ["name", "target"]) {
    if (typeof stage[field] !== "string" || stage[field].trim() === "") {
      throw new Error(`${label} must declare ${field}`);
    }
  }
  if (stage.children !== undefined) {
    throw new Error(`browser E2E batch stage ${stage.name} must use summary_children[], not legacy children[]`);
  }

  const summaryChildren = stage.summary_children ?? [];
  if (!Array.isArray(summaryChildren)) {
    throw new Error(`browser E2E batch stage ${stage.name} summary_children must be an array`);
  }
  const normalizedSummaryChildren = summaryChildren.map((child, childIndex) => {
    if (typeof child !== "string" || child.trim() === "") {
      throw new Error(
        `browser E2E batch stage ${stage.name} summary_children ${childIndex + 1} must be a non-empty string`,
      );
    }
    return child.trim();
  });
  const duplicateSummaryChild = normalizedSummaryChildren.find(
    (child, childIndex, children) => children.indexOf(child) !== childIndex,
  );
  if (duplicateSummaryChild) {
    throw new Error(
      `browser E2E batch stage ${stage.name} summary_children contains duplicate ${duplicateSummaryChild}`,
    );
  }

  const groups = stage.groups ?? [];
  if (!Array.isArray(groups) || groups.length === 0) {
    throw new Error(`browser E2E batch stage ${stage.name} must declare groups[]`);
  }
  const normalizedGroups = groups.map((group, groupIndex) =>
    normalizeGroup(stage.name, group, groupIndex + 1, runtimeProfiles),
  );
  const duplicateGroup = normalizedGroups.find(
    (group, groupIndex) => normalizedGroups.findIndex((candidate) => candidate.name === group.name) !== groupIndex,
  );
  if (duplicateGroup) {
    throw new Error(`browser E2E batch stage ${stage.name} contains duplicate group ${duplicateGroup.name}`);
  }
  const groupTargets = new Set(normalizedGroups.map((group) => group.target));
  for (const child of normalizedSummaryChildren) {
    if (!groupTargets.has(child)) {
      throw new Error(`browser E2E batch stage ${stage.name} summary child ${child} must match a group target`);
    }
  }

  return {
    name: stage.name.trim(),
    target: stage.target.trim(),
    summaryChildren: normalizedSummaryChildren,
    scheduleTags: normalizeScheduleTags(stage),
    schedulerNeeds: normalizeSchedulerNeeds(stage),
    groups: normalizedGroups,
  };
}

function normalizeScheduleTags(stage) {
  if (stage.schedule_tags === undefined) {
    return [];
  }
  if (!Array.isArray(stage.schedule_tags)) {
    throw new Error(`browser E2E batch stage ${stage.name} schedule_tags must be an array`);
  }
  const tags = [];
  const seen = new Set();
  for (const [index, tag] of stage.schedule_tags.entries()) {
    if (typeof tag !== "string" || !/^[a-z][a-z0-9_]*$/.test(tag)) {
      throw new Error(
        `browser E2E batch stage ${stage.name} schedule_tags ${index + 1} must be a snake_case token`,
      );
    }
    if (seen.has(tag)) {
      throw new Error(`browser E2E batch stage ${stage.name} schedule_tags contains duplicate ${tag}`);
    }
    seen.add(tag);
    tags.push(tag);
  }
  return tags;
}

function normalizeSchedulerNeeds(stage) {
  if (stage.scheduler_dependency_policy !== undefined) {
    throw new Error(`browser E2E batch stage ${stage.name} must use scheduler_needs[], not obsolete scheduler_dependency_policy`);
  }
  if (stage.scheduler_needs === undefined) {
    return [];
  }
  if (!Array.isArray(stage.scheduler_needs)) {
    throw new Error(`browser E2E batch stage ${stage.name} scheduler_needs must be an array`);
  }
  const needs = [];
  const seen = new Set();
  for (const [index, need] of stage.scheduler_needs.entries()) {
    if (typeof need !== "string" || need.trim() === "") {
      throw new Error(
        `browser E2E batch stage ${stage.name} scheduler_needs ${index + 1} must be a non-empty string`,
      );
    }
    const normalized = need.trim();
    if (seen.has(normalized)) {
      throw new Error(`browser E2E batch stage ${stage.name} scheduler_needs contains duplicate ${normalized}`);
    }
    seen.add(normalized);
    needs.push(normalized);
  }
  return needs;
}

function normalizeGroup(stageName, group, index, runtimeProfiles) {
  if (!group || typeof group !== "object" || Array.isArray(group)) {
    throw new Error(`browser E2E batch stage ${stageName} group ${index} must be an object`);
  }
  for (const key of ["name", "target", "kind"]) {
    if (typeof group[key] !== "string" || group[key].trim() === "") {
      throw new Error(`browser E2E batch stage ${stageName} group ${index} must declare ${key}`);
    }
  }
  if (!allowedGroupKinds.has(group.kind)) {
    throw new Error(`browser E2E batch group ${group.name} has unsupported kind ${group.kind}`);
  }
  const runtimeProfileID = normalizeOptionalString(group.runtime_profile_id) || "default";
  if (!runtimeProfiles.has(runtimeProfileID)) {
    throw new Error(
      `browser E2E batch group ${group.name} references unknown runtime profile ${runtimeProfileID}`,
    );
  }
  const resourceProfileID = normalizeOptionalString(group.resource_profile_id);
  if (!resourceProfileID) {
    throw new Error(`browser E2E batch group ${group.name} must declare resource_profile_id`);
  }
  if (group.kind === "measurement" && resourceProfileID !== "browser_measurement_quiet") {
    throw new Error(`browser E2E batch measurement group ${group.name} must use browser_measurement_quiet`);
  }
  if (group.kind !== "measurement" && resourceProfileID === "browser_measurement_quiet") {
    throw new Error(`browser E2E batch non-measurement group ${group.name} cannot use browser_measurement_quiet`);
  }
  const serviceRequirement = normalizeOptionalString(group.service_requirement);
  if (!new Set(["none", "test-services"]).has(serviceRequirement)) {
    throw new Error(
      `browser E2E batch group ${group.name} must declare service_requirement none or test-services`,
    );
  }
  if (serviceRequirement !== runtimeProfiles.get(runtimeProfileID)) {
    throw new Error(
      `browser E2E batch group ${group.name} service requirement ${serviceRequirement} does not match runtime profile ${runtimeProfileID}`,
    );
  }
  if (!Array.isArray(group.service_dependencies)) {
    throw new Error(`browser E2E batch group ${group.name} must declare service_dependencies`);
  }
  const serviceDependencies = group.service_dependencies.map((entry) => String(entry));
  if (
    JSON.stringify(serviceDependencies) !== JSON.stringify([...serviceDependencies].sort()) ||
    new Set(serviceDependencies).size !== serviceDependencies.length ||
    serviceDependencies.some((entry) => !new Set(["object_store", "postgres"]).has(entry))
  ) {
    throw new Error(`browser E2E batch group ${group.name} has invalid service_dependencies`);
  }
  const browserSessionGroup = normalizeOptionalString(group.browser_session_group);
  if (!browserSessionGroup) {
    throw new Error(`browser E2E batch group ${group.name} must declare browser_session_group`);
  }
  const browserSessionIsolationReason = normalizeOptionalString(group.browser_session_isolation_reason);
  if (runtimeProfileID !== "default" && !browserSessionIsolationReason) {
    throw new Error(
      `browser E2E batch group ${group.name} must explain non-default runtime-profile session isolation`,
    );
  }
  return {
    name: group.name.trim(),
    target: group.target.trim(),
    kind: group.kind.trim(),
    coverage: normalizeCoverage(group),
    executionDependency:
      group.execution_dependency === undefined ? "" : String(group.execution_dependency).trim(),
    workers: group.workers === undefined ? "default" : String(group.workers),
    resetBefore: group.reset_before === undefined ? "" : String(group.reset_before),
    selectedRowIDs: normalizeSelectedRowIDs(group),
    specs: normalizeSpecs(group),
    browserSessionGroup,
    browserSessionIsolationReason,
    runtimeProfileID,
    resourceProfileID,
    serviceDependencies,
    serviceRequirement,
  };
}

function normalizeSpecs(group) {
  if (!Array.isArray(group.specs) || group.specs.length !== 1) {
    throw new Error(`browser E2E batch group ${group.name} specs must contain exactly one catalog selector file`);
  }
  const file = String(group.specs[0] ?? "").trim();
  if (!/^apps\/web\/e2e\/.+\.spec\.ts$/u.test(file) || file.includes("..")) {
    throw new Error(`browser E2E batch group ${group.name} has an unsafe or unsupported spec path`);
  }
  return [file];
}

function normalizeOptionalString(value) {
  return value === undefined ? "" : String(value).trim();
}

function normalizeSelectedRowIDs(group) {
  if (group.selected_row_ids === undefined) {
    throw new Error(`browser E2E batch group ${group.name} must declare selected_row_ids`);
  }
  if (!Array.isArray(group.selected_row_ids)) {
    throw new Error(`browser E2E batch group ${group.name} selected_row_ids must be an array`);
  }
  const ids = [];
  const seen = new Set();
  for (const [index, raw] of group.selected_row_ids.entries()) {
    const id = String(raw ?? "").trim();
    if (id === "") {
      throw new Error(`browser E2E batch group ${group.name} selected_row_ids ${index + 1} must be non-empty`);
    }
    if (!/^[a-z][a-z0-9_]*(?:\.[a-z0-9_]+)+$/u.test(id)) {
      throw new Error(`browser E2E batch group ${group.name} selected_row_ids contains non-semantic identity ${id}`);
    }
    if (seen.has(id)) {
      throw new Error(`browser E2E batch group ${group.name} selected_row_ids contains duplicate ${id}`);
    }
    seen.add(id);
    ids.push(id);
  }
  if (ids.length === 0) {
    throw new Error(`browser E2E batch group ${group.name} selected_row_ids must not be empty`);
  }
  return ids;
}

function normalizeCoverage(group) {
  if (group.coverage === undefined) {
    return group.kind === "support" ? "supplemental" : "";
  }
  const coverage = String(group.coverage).trim();
  if (!allowedCoverage.has(coverage)) {
    throw new Error(`browser E2E batch group ${group.name} has unsupported coverage ${coverage}`);
  }
  return coverage;
}

function printTargetMetadata(stage) {
  process.stdout.write(`${stage.target}\n`);
  process.stdout.write(`${stage.summaryChildren.join(",")}\n`);
}

function printRunnerMetadata(stage) {
  printTargetMetadata(stage);
  for (const group of stage.groups) {
    process.stdout.write(
      [
        group.name,
        group.target,
        group.kind,
        group.workers,
        group.resetBefore,
        group.coverage,
        group.executionDependency,
        stage.scheduleTags.join(","),
        stage.schedulerNeeds.join(","),
        group.selectedRowIDs.join(","),
        group.browserSessionGroup,
        group.browserSessionIsolationReason,
        group.runtimeProfileID,
        group.serviceRequirement,
        group.specs.join(","),
      ].join("\t") + "\n",
    );
  }
}

function printGroupSelections(manifestPath) {
  const stages = loadBrowserBatchStages(manifestPath);
  for (const stage of stages.values()) {
    for (const group of stage.groups) {
      process.stdout.write(
        [
          stage.name,
          group.name,
          group.target,
          group.kind,
          group.coverage,
          group.executionDependency,
          group.selectedRowIDs.join(","),
        ].join("\t") + "\n",
      );
    }
  }
}

function printStageSessions(stage) {
  const sessions = new Map();
  for (const group of stage.groups) {
    const sessionGroup = group.browserSessionGroup || stage.target;
    const current = sessions.get(sessionGroup) ?? {
      runtimeProfileID: group.runtimeProfileID,
      groupNames: [],
    };
    if (current.runtimeProfileID !== group.runtimeProfileID) {
      throw new Error(
        `browser session ${sessionGroup} mixes runtime profiles ${current.runtimeProfileID} and ${group.runtimeProfileID}`,
      );
    }
    current.groupNames.push(group.name);
    sessions.set(sessionGroup, current);
  }
  for (const [sessionGroup, session] of sessions) {
    process.stdout.write(
      [sessionGroup, session.runtimeProfileID, session.groupNames.join(",")].join("\t") + "\n",
    );
  }
}

function main(argv) {
  const [command, manifestPath, stageName] = argv;
  switch (command) {
    case "validate":
      if (!manifestPath || stageName !== undefined) {
        throw new Error("usage: browser-batch-manifest.mjs validate <manifest>");
      }
      loadBrowserBatchStages(manifestPath);
      return;
    case "stage-target":
      if (!manifestPath || !stageName) {
        throw new Error("usage: browser-batch-manifest.mjs stage-target <manifest> <stage>");
      }
      printTargetMetadata(resolveBrowserBatchStage(manifestPath, stageName));
      return;
    case "stage-runner":
      if (!manifestPath || !stageName) {
        throw new Error("usage: browser-batch-manifest.mjs stage-runner <manifest> <stage>");
      }
      printRunnerMetadata(resolveBrowserBatchStage(manifestPath, stageName));
      return;
    case "group-selections":
      if (!manifestPath || stageName !== undefined) {
        throw new Error("usage: browser-batch-manifest.mjs group-selections <manifest>");
      }
      printGroupSelections(manifestPath);
      return;
    case "stage-sessions":
      if (!manifestPath || !stageName) {
        throw new Error("usage: browser-batch-manifest.mjs stage-sessions <manifest> <stage>");
      }
      printStageSessions(resolveBrowserBatchStage(manifestPath, stageName));
      return;
    default:
      throw new Error(
        "usage: browser-batch-manifest.mjs <validate|stage-target|stage-runner|stage-sessions|group-selections> <manifest> [stage]",
      );
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exit(1);
  }
}
