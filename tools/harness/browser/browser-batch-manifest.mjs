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

export const browserBatchManifestSchemaID = "cartulary.browser_e2e_batch_manifest.v5";

const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const browserBatchKeys = new Set(["schema_id", "stages"]);
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
  "specs",
]);
const allowedGroupKinds = new Set([
  "webserver-backed",
  "duration_balanced_specs",
  "functional",
  "support",
  "stateful",
  "measurement",
  "visual",
  "a11y",
  "a11y_preflight",
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
  const stages = new Map();
  for (const [index, rawStage] of manifest.stages.entries()) {
    const stage = normalizeStage(rawStage, index + 1);
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

function normalizeStage(stage, index) {
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
  const normalizedGroups = groups.map((group, groupIndex) => normalizeGroup(stage.name, group, groupIndex + 1));
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

function normalizeGroup(stageName, group, index) {
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
  return {
    name: group.name.trim(),
    target: group.target.trim(),
    kind: group.kind.trim(),
    coverage: normalizeCoverage(group),
    executionDependency:
      group.execution_dependency === undefined ? "" : String(group.execution_dependency).trim(),
    workers: group.workers === undefined ? "default" : String(group.workers),
    resetBefore: group.reset_before === undefined ? "" : String(group.reset_before),
  };
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
        ].join("\t") + "\n",
      );
    }
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
    default:
      throw new Error(
        "usage: browser-batch-manifest.mjs <validate|stage-target|stage-runner|group-selections> <manifest> [stage]",
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
