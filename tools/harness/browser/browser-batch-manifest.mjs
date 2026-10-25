#!/usr/bin/env node
import { readFileSync } from "node:fs";
import { validateSchemaSync } from "../contract/harness-contract.mjs";

import {
  readJsonObject,
  requireObject,
  requireSchemaID,
} from "../contract/json-shape.mjs";

import { browserBatchManifestSchemaID, browserPartitionContract } from "./browser-compatibility.mjs";
export { browserBatchManifestSchemaID };

const allowedGroupKinds = new Set([
  "webserver-backed",
  "duration_balanced_specs",
  "support",
  "stateful",
  "stateful_partition",
  "measurement",
  "visual",
  "a11y",
]);
function manifestValue(fileOrManifest, label) {
  return typeof fileOrManifest === "string"
    ? readJsonObject(fileOrManifest, label)
    : requireObject(fileOrManifest, label);
}

export function validateBrowserBatchManifestShape(fileOrManifest, label = fileOrManifest) {
  const manifest = manifestValue(fileOrManifest, label);
  requireSchemaID(manifest, browserBatchManifestSchemaID, label);
  validateSchemaSync(browserBatchManifestSchemaID, manifest);
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
  validateBrowserBatchManifestShape(manifest, "browser batch manifest");
  const runtimeProfiles = new Map(
    (manifest.runtime_profiles ?? []).map((profile) => [
      profile.id,
      profile,
    ]),
  );
  if (!runtimeProfiles.has("default")) {
    throw new Error("browser E2E batch manifest must declare the default runtime profile");
  }
  const partitions = normalizePartitions(manifest, runtimeProfiles);
  const stages = new Map();
  const covered = new Set();
  for (const [index, rawStage] of manifest.stages.entries()) {
    const groups = rawStage.groups.map((membership) => {
      const partition = partitions.get(membership.partition);
      if (!partition) throw new Error(`unknown browser partition ${membership.partition}`);
      for (const id of membership.selected_row_ids) {
        if (!partition.selectedRowIDs.includes(id) || covered.has(id)) throw new Error(`invalid or duplicate browser reporting membership ${id}`);
        covered.add(id);
      }
      return { ...partition, selectedRowIDs: [...membership.selected_row_ids] };
    });
    if (stages.has(rawStage.name)) throw new Error(`duplicate browser batch stage ${rawStage.name}`);
    stages.set(rawStage.name, { name: rawStage.name, target: rawStage.target,
      summaryChildren: rawStage.summary_children ?? [], scheduleTags: normalizeScheduleTags(rawStage),
      schedulerNeeds: normalizeSchedulerNeeds(rawStage), groups, invocationPolicy: manifest.invocation_policy });
  }
  if ([...partitions.values()].some((partition) => partition.selectedRowIDs.some((id) => !covered.has(id)))) {
    throw new Error("browser partition lacks exact reporting membership");
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

function normalizePartitions(manifest, runtimeProfiles) {
  const partitions = new Map(), rows = new Set();
  for (const [index, raw] of manifest.partitions.entries()) {
    const group = normalizeGroup({ name: raw.execution_stage, target: raw.target,
      frontend_artifact_id: raw.frontend_artifact_id, scheduler_needs: raw.scheduler_needs }, raw, index + 1, runtimeProfiles);
    if (partitions.has(group.name)) throw new Error(`duplicate browser partition ${group.name}`);
    for (const row of group.selectedRowIDs) {
      if (rows.has(row)) throw new Error(`duplicate browser execution row ${row}`);
      rows.add(row);
    }
    partitions.set(group.name, group);
  }
  return partitions;
}

export function resolveBrowserBatchPartition(manifestPath, name) {
  const manifest = loadBrowserBatchManifest(manifestPath);
  // Validate reporting membership as well as the execution contract.
  normalizeBrowserBatchStages(manifest);
  const partitions = normalizePartitions(manifest, new Map(manifest.runtime_profiles.map((profile) => [profile.id, profile])));
  const partition = partitions.get(name);
  if (!partition) throw new Error(`unknown browser partition ${name}`);
  return { partition, invocationPolicy: manifest.invocation_policy };
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

function normalizeGroup(stage, group, index, runtimeProfiles) {
  const stageName = stage.name;
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
  const runtimeProfileID = group.runtime_profile_id;
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
  if (serviceRequirement !== runtimeProfiles.get(runtimeProfileID).service_requirement) {
    throw new Error(
      `browser E2E batch group ${group.name} service requirement ${serviceRequirement} does not match runtime profile ${runtimeProfileID}`,
    );
  }
  if (!Array.isArray(group.service_dependencies)) {
    throw new Error(`browser E2E batch group ${group.name} must declare service_dependencies`);
  }
  if (group.postgres_isolation!=="exclusive") throw new Error(`${group.name} requires exclusive PostgreSQL isolation`);
  const serviceDependencies = group.service_dependencies.map((entry) => String(entry));
  if (
    JSON.stringify(serviceDependencies) !== JSON.stringify([...serviceDependencies].sort()) ||
    new Set(serviceDependencies).size !== serviceDependencies.length ||
    serviceDependencies.some((entry) => !new Set(["object_store", "postgres"]).has(entry))
  ) {
    throw new Error(`browser E2E batch group ${group.name} has invalid service_dependencies`);
  }
  if (!Array.isArray(group.preparation_operations) ||
      group.preparation_operations.some((id) => typeof id !== "string" || !/^harness\.[a-z][a-z0-9_]*$/u.test(id)) ||
      JSON.stringify(group.preparation_operations) !== JSON.stringify([...new Set(group.preparation_operations)].sort())) {
    throw new Error(`browser partition ${group.name} requires exact preparation operations`);
  }
  if (group.workers !== "1") throw new Error(`browser E2E batch group ${group.name} requires exactly one worker`);
  const browserSessionGroup = normalizeOptionalString(group.browser_session_group);
  if (!browserSessionGroup) {
    throw new Error(`browser E2E batch group ${group.name} must declare browser_session_group`);
  }
  if (!["test_owned", "isolated", "measurement"].includes(group.state_scope) ||
      group.fixture_capability !== "browser_stack" ||
      typeof group.project !== "string" || !group.project || group.config !== "playwright.config.ts") {
    throw new Error(`browser partition ${group.name} requires browser fixture capability, project, config and explicit state scope`);
  }
  if (group.state_scope !== "test_owned" && group.selected_row_ids.length !== 1) {
    throw new Error(`browser partition ${group.name} requires singleton isolation`);
  }
  return {
    ...browserPartitionContract(stage, group, runtimeProfiles.get(runtimeProfileID)),
    name: group.name.trim(),
    selectedRowIDs: normalizeSelectedRowIDs(group),
    specs: normalizeSpecs(group),
  };
}

function normalizeSpecs(group) {
  if (!Array.isArray(group.specs) || group.specs.length === 0 ||
      new Set(group.specs).size !== group.specs.length ||
      JSON.stringify(group.specs) !== JSON.stringify([...group.specs].sort()) ||
      group.specs.some((file) => typeof file !== "string" ||
        !/^apps\/web\/e2e\/.+\.spec\.ts$/u.test(file) || file.includes(".."))) {
    throw new Error(`browser partition ${group.name} requires sorted unique catalog selector files`);
  }
  return [...group.specs];
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
