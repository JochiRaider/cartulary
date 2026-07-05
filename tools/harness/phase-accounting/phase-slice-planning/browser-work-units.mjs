import path from "node:path";

import { normalizeBrowserBatchStages } from "../../browser/browser-batch-manifest.mjs";
import { loadExecutionTopology, renderBrowserBatchManifest } from "../../generated-artifacts/execution-topology.mjs";
import { browserStageResource } from "../../scheduler/scheduler-resources.mjs";
import {
  disabledFrontendRowAccountingScope,
  executionDependenciesForTarget,
} from "./row-selection.mjs";
import {
  browserStackResource,
  targetWeight,
  uniqueSorted,
} from "./work-unit-common.mjs";

const preferredBrowserStageByDependency = new Map([
  ["browser_functional", "webserver-backed"],
  ["browser_stateful", "stateful"],
  ["browser_measurement", "measurement"],
  ["browser_visual", "visual"],
  ["browser_a11y", "a11y"],
  ["browser_a11y_preflight", "a11y-preflight"],
]);

export function resolveBrowserStagesByTarget(root) {
  const topology = loadExecutionTopology({
    manifestPath: path.join(root, "tools", "execution_topology_manifest.json"),
  });
  const stages = normalizeBrowserBatchStages(renderBrowserBatchManifest(topology));
  const byTarget = new Map();
  for (const stage of stages.values()) {
    if (!byTarget.has(stage.target)) {
      byTarget.set(stage.target, []);
    }
    byTarget.get(stage.target).push(stage);
  }
  return byTarget;
}

function browserStageDependencies(stage) {
  return uniqueSorted(
    stage.groups
      .filter((group) => group.coverage !== "raw")
      .map((group) => group.executionDependency)
      .filter(Boolean),
  );
}

function resolveBrowserStageForRows(target, rows, stageByTarget) {
  const candidates = stageByTarget.get(target) ?? [];
  if (candidates.length === 0) {
    throw new Error(`phase slice browser target ${target} is not a browser batch stage target`);
  }
  if (candidates.length === 1) {
    return candidates[0];
  }

  const dependencies = executionDependenciesForTarget(rows, target);
  const preferredNames = uniqueSorted(
    dependencies.map((dependency) => preferredBrowserStageByDependency.get(dependency)),
  );
  const preferredCandidates = candidates.filter((stage) => preferredNames.includes(stage.name));
  if (preferredCandidates.length === 1) {
    return preferredCandidates[0];
  }

  const matchingCandidates = candidates.filter((stage) => {
    const stageDependencies = new Set(browserStageDependencies(stage));
    return dependencies.every((dependency) => stageDependencies.has(dependency));
  });
  if (matchingCandidates.length === 1) {
    return matchingCandidates[0];
  }

  throw new Error(
    `phase slice browser target ${target} matches multiple browser batch stages; declare an explicit dependency-to-stage selector`,
  );
}

function browserNeeds(plan, stage) {
  const selectedTargets = new Set(plan.child_target_names);
  const needs = stage.schedulerNeeds ?? [];
  for (const need of needs) {
    if (need === stage.target) {
      throw new Error(`phase slice browser target ${stage.target} must not depend on itself`);
    }
    if (!selectedTargets.has(need)) {
      throw new Error(`phase slice browser target ${stage.target} scheduler_needs target ${need} is not selected by the slice`);
    }
  }
  return needs;
}

export function addBrowserUnit(plan, target, rows, stageByTarget) {
  const stage = resolveBrowserStageForRows(target, rows, stageByTarget);
  const claims = new Map([
    ["postgres", 1],
    ["object_store", 1],
    ["process", 1],
    [browserStackResource, 1],
    [browserStageResource(stage.name), 1],
  ]);
  plan.browserStages.add(stage.name);
  plan.workUnits.push({
    id: target,
    label: target,
    kind: "browser_target",
    type: "make_target",
    class: "browser",
    target,
    aggregateTarget: target,
    group: target,
    browserStage: stage.name,
    needs: browserNeeds(plan, stage),
    completionKeys: [target],
    failureKeys: [target],
    weightMs: targetWeight(rows),
    resourceClaims: claims,
    frontend_row_accounting_scope: disabledFrontendRowAccountingScope(plan.phase),
    order: plan.nextOrder++,
  });
}
