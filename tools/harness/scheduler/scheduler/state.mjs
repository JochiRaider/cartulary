import {
  schedulerBlockedUnitRecords,
  schedulerProgressIntervalMs,
  schedulerWaitingOnForUnits,
} from "../scheduler-reporting.mjs";
import { formatResourceMap } from "../scheduler-resources.mjs";

export function progressDelay() {
  let timeout;
  const promise = new Promise((resolve) => {
    timeout = setTimeout(() => resolve({ schedulerProgressTick: true }), schedulerProgressIntervalMs);
  });
  return {
    promise,
    cancel() {
      clearTimeout(timeout);
    },
  };
}

export function counted(unit) {
  return unit.countInTotal !== false;
}

export function finalizer(unit) {
  return (
    unit.kind === "finalizer" ||
    unit.kind === "aggregate_finalize" ||
    unit.kind === "browser_session_finalizer" ||
    unit.browserSessionFinalizer === true
  );
}

export function visibleRunningCount(running) {
  return Array.from(running.values()).filter(counted).length;
}

export function visiblePendingCount(pending) {
  return pending.filter(counted).length;
}

export function pendingFinalizerCount(pending) {
  return pending.filter(finalizer).length;
}

export function runningFinalizerCount(running) {
  return Array.from(running.values()).filter(finalizer).length;
}

export function unitCompletionKeys(unit) {
  return unit.completionKeys?.length ? unit.completionKeys : [unit.id];
}

export function unitFailureKeys(unit) {
  return unit.failureKeys?.length ? unit.failureKeys : unitCompletionKeys(unit);
}

function hasFailedDependency(unit, failedKeys) {
  if (finalizer(unit)) return null;
  return (unit.needs ?? []).find((need) => failedKeys.has(need)) ?? null;
}

function dependenciesSatisfied(unit, completedKeys, failedKeys) {
  return (unit.needs ?? []).every((need) =>
    completedKeys.has(need) || (finalizer(unit) && failedKeys.has(need)),
  );
}

function readyPendingUnits(pending, completedKeys, failedKeys) {
  return pending.filter(
    (unit) => !hasFailedDependency(unit, failedKeys) && dependenciesSatisfied(unit, completedKeys, failedKeys),
  );
}

function dependencyBlockedPendingUnits(pending, completedKeys, failedKeys) {
  return pending.filter(
    (unit) => counted(unit) && !hasFailedDependency(unit, failedKeys) && !dependenciesSatisfied(unit, completedKeys, failedKeys),
  );
}

function hasResourceCapacity(unit, resourceLimits, activeClaims) {
  return blockedResourcesForUnit(unit, resourceLimits, activeClaims).length === 0;
}

function claimsReservedResource(unit, reservedResources) {
  for (const resource of unit.resourceClaims.keys()) {
    if (reservedResources.has(resource)) {
      return true;
    }
  }
  return false;
}

export function priorityAdmissiblePendingUnitIndex({
  pending,
  completedKeys,
  failedKeys,
  resourceLimits,
  activeClaims,
}) {
  const reservedResources = new Set();
  for (const [index, candidate] of pending.entries()) {
    if (hasFailedDependency(candidate, failedKeys) || !dependenciesSatisfied(candidate, completedKeys, failedKeys)) {
      continue;
    }
    const blockedResources = blockedResourcesForUnit(candidate, resourceLimits, activeClaims);
    if (blockedResources.length === 0) {
      if (!claimsReservedResource(candidate, reservedResources)) {
        return index;
      }
      continue;
    }
    for (const resource of blockedResources) {
      reservedResources.add(resource);
    }
  }
  return -1;
}

function blockedResourcesForUnit(unit, resourceLimits, activeClaims) {
  const blocked = [];
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    const limit = resourceLimits.get(resource);
    if (limit !== undefined && (activeClaims.get(resource) ?? 0) + amount > limit) {
      blocked.push(resource);
    }
  }
  return blocked.sort((left, right) => left.localeCompare(right));
}

function blockedResourcesForUnits(units, resourceLimits, activeClaims) {
  const resources = new Set();
  for (const unit of units) {
    for (const resource of blockedResourcesForUnit(unit, resourceLimits, activeClaims)) {
      resources.add(resource);
    }
  }
  return Array.from(resources).sort((left, right) => left.localeCompare(right));
}

export function addResourceClaims(unit, activeClaims) {
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    activeClaims.set(resource, (activeClaims.get(resource) ?? 0) + amount);
  }
}

export function removeResourceClaims(unit, activeClaims) {
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    const next = (activeClaims.get(resource) ?? 0) - amount;
    if (next <= 0) {
      activeClaims.delete(resource);
    } else {
      activeClaims.set(resource, next);
    }
  }
}

export function formatBlockedWorkUnits(workUnits) {
  return workUnits
    .map((unit) => `${unit.label} claims=${formatResourceMap(unit.resourceClaims)}`)
    .join("; ");
}

export function stateFields({ pending, running, activeClaims, resourceLimits }) {
  return [
    `active=${visibleRunningCount(running)}`,
    `pending=${visiblePendingCount(pending)}`,
    `active_resource_claims=${formatResourceMap(activeClaims)}`,
    `resource_limits=${formatResourceMap(resourceLimits)}`,
  ];
}

export function skippedReasonForStoppedUnit(unit, completedKeys, failedKey, unitsByCompletionKey, memo = new Map()) {
  if (memo.has(unit.id)) {
    return memo.get(unit.id);
  }
  for (const need of unit.needs ?? []) {
    if (need === failedKey) {
      memo.set(unit.id, "dependency_failure");
      return "dependency_failure";
    }
    if (!completedKeys.has(need)) {
      const upstream = unitsByCompletionKey.get(need);
      if (
        upstream &&
        skippedReasonForStoppedUnit(upstream, completedKeys, failedKey, unitsByCompletionKey, memo) ===
          "dependency_failure"
      ) {
        memo.set(unit.id, "dependency_failure");
        return "dependency_failure";
      }
    }
  }
  memo.set(unit.id, "schedule_stopped_after_failure");
  return "schedule_stopped_after_failure";
}

function failedDependencyForSkip(unit, failedKeys) {
  if (finalizer(unit)) return null;
  for (const need of unit.needs ?? []) {
    if (failedKeys.has(need)) {
      return need;
    }
  }
  return null;
}

export function skipFailedDependencyUnits({ pending, failedKeys, reporter, stateSnapshot }) {
  let skipped = 0;
  let skippedThisPass = 0;
  do {
    skippedThisPass = 0;
    for (let index = 0; index < pending.length; ) {
      const unit = pending[index];
      const failedDependency = failedDependencyForSkip(unit, failedKeys);
      if (!failedDependency) {
        index += 1;
        continue;
      }
      pending.splice(index, 1);
      skipped += 1;
      skippedThisPass += 1;
      for (const key of unitFailureKeys(unit)) {
        failedKeys.set(key, failedDependency);
      }
      reporter.skipUnit(
        unit,
        stateSnapshot(skipped),
        "dependency_failure",
        failedDependency,
      );
    }
  } while (skippedThisPass > 0);
  return skipped;
}

export function blockedSnapshot({ pending, completedKeys, failedKeys, resourceLimits, activeClaims }) {
  const dependencyBlocked = dependencyBlockedPendingUnits(pending, completedKeys, failedKeys);
  const readyBlocked = readyPendingUnits(pending, completedKeys, failedKeys)
    .filter(counted)
    .filter((unit) => !hasResourceCapacity(unit, resourceLimits, activeClaims));
  const blockedResources = blockedResourcesForUnits(readyBlocked, resourceLimits, activeClaims);
  const waitingOn = schedulerWaitingOnForUnits(dependencyBlocked, completedKeys);
  const blockedUnits = schedulerBlockedUnitRecords({
    dependencyBlocked,
    resourceBlocked: readyBlocked,
    completed: completedKeys,
    blockedResourcesForUnit: (unit) => blockedResourcesForUnit(unit, resourceLimits, activeClaims),
  });
  let reason = "none";
  if (dependencyBlocked.length > 0 && readyBlocked.length > 0) {
    reason = "dependencies,resources";
  } else if (dependencyBlocked.length > 0) {
    reason = "dependencies";
  } else if (readyBlocked.length > 0) {
    reason = "resources";
  }
  return {
    dependencyBlocked,
    readyBlocked,
    blockedResources,
    waitingOn,
    blockedUnits,
    blockedCount: dependencyBlocked.length + readyBlocked.length,
    reason,
  };
}
