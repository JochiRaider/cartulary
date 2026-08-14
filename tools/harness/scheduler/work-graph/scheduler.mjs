import { validateSchemaSync } from "../../contract/index.mjs";
import { validateWorkGraph } from "./model.mjs";
import { executeUnitProcess } from "./executor.mjs";

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function successorsByID(graph) {
  const successors = new Map(graph.units.map((unit) => [unit.unit_id, []]));
  for (const unit of graph.units) {
    for (const dependency of unit.needs) successors.get(dependency).push(unit.unit_id);
  }
  for (const values of successors.values()) values.sort(compareASCII);
  return successors;
}

export function criticalPathRanks(graph) {
  const byID = new Map(graph.units.map((unit) => [unit.unit_id, unit]));
  const successors = successorsByID(graph);
  const memo = new Map();
  function rank(unitID) {
    if (memo.has(unitID)) return memo.get(unitID);
    const downstream = successors.get(unitID).map(rank);
    const value = byID.get(unitID).estimated_work_ms + Math.max(0, ...downstream);
    memo.set(unitID, value);
    return value;
  }
  for (const unit of graph.units) rank(unit.unit_id);
  return memo;
}

function fitting(unit, capacities, activeClaims, activeSharedLocks, activeExclusiveLocks) {
  const resourcesFit = Object.entries(unit.resource_claims).every(
    ([resource, amount]) =>
      (activeClaims.get(resource) ?? 0) + amount <= capacities.get(resource),
  );
  if (!resourcesFit) return false;
  if ((unit.shared_locks ?? []).some((lock) => activeExclusiveLocks.has(lock))) return false;
  return !(unit.exclusive_locks ?? []).some(
    (lock) => activeExclusiveLocks.has(lock) || (activeSharedLocks.get(lock) ?? 0) > 0,
  );
}

function blockedByReadyExclusiveWaiter(unit, readyUnits) {
  return (unit.shared_locks ?? []).some((lock) =>
    readyUnits.some((waiting) =>
      waiting.unit_id !== unit.unit_id && (waiting.exclusive_locks ?? []).includes(lock),
    ),
  );
}

function blockingDetails(unit, capacities, activeClaims, activeSharedLocks, activeExclusiveLocks, running, byID) {
  const resources = Object.entries(unit.resource_claims)
    .filter(([resource, amount]) => (activeClaims.get(resource) ?? 0) + amount > capacities.get(resource))
    .map(([resource]) => resource);
  const conflictingLocks = [
    ...(unit.shared_locks ?? []).filter((lock) => activeExclusiveLocks.has(lock)),
    ...(unit.exclusive_locks ?? []).filter(
      (lock) => activeExclusiveLocks.has(lock) || (activeSharedLocks.get(lock) ?? 0) > 0,
    ),
  ];
  const holders = [...running.keys()].filter((unitID) => {
    const active = byID.get(unitID);
    return resources.some((resource) => active.resource_claims[resource]) ||
      conflictingLocks.some((lock) =>
        (active.shared_locks ?? []).includes(lock) || (active.exclusive_locks ?? []).includes(lock),
      );
  });
  return {
    blocking_resources: [
      ...resources,
      ...conflictingLocks.map((lock) => `lock_${lock.replaceAll(/[^a-zA-Z0-9_]+/gu, "_")}`),
    ].sort(compareASCII),
    resource_holders: holders.sort(compareASCII),
  };
}

function addClaims(activeClaims, claims, direction) {
  for (const [resource, amount] of Object.entries(claims)) {
    const next = (activeClaims.get(resource) ?? 0) + direction * amount;
    if (next === 0) activeClaims.delete(resource);
    else activeClaims.set(resource, next);
  }
}

function addLocks(activeSharedLocks, activeExclusiveLocks, unit, direction) {
  for (const lock of unit.shared_locks ?? []) {
    const next = (activeSharedLocks.get(lock) ?? 0) + direction;
    if (next === 0) activeSharedLocks.delete(lock);
    else activeSharedLocks.set(lock, next);
  }
  for (const lock of unit.exclusive_locks ?? []) {
    if (direction > 0) activeExclusiveLocks.add(lock);
    else activeExclusiveLocks.delete(lock);
  }
}

const dependencyFailureStates = new Set(["failed", "skipped", "cancelled"]);

function dependencyBlocksDescendants(dependency, state, byID) {
  return dependencyFailureStates.has(state.get(dependency)) &&
    byID.get(dependency)?.failure_policy.block_descendants !== false;
}

function failedDependency(unit, state, byID) {
  return unit.needs.find((dependency) =>
    dependencyBlocksDescendants(dependency, state, byID),
  );
}

function ready(unit, state, byID) {
  return unit.needs.every((dependency) =>
    state.get(dependency) === "passed" ||
    (dependencyFailureStates.has(state.get(dependency)) &&
      byID.get(dependency)?.failure_policy.block_descendants === false),
  );
}

function schedulerEvent(seq, monotonicMs, event, unit, status, extra = {}) {
  const record = {
    schema_id: "cartulary.harness_unit_event.v1",
    seq,
    monotonic_ms: monotonicMs,
    event,
    unit_id: unit.unit_id,
    status,
    resource_claims: unit.resource_claims,
    service_dependencies: unit.service_dependencies ?? [],
    ...extra,
  };
  validateSchemaSync(record.schema_id, record);
  return record;
}

export function simulateWorkGraph({
  graph,
  capacities,
  durations = new Map(),
  outcomes = new Map(),
  agingQuantumMs = 1000,
  cancelAtMs,
}) {
  validateWorkGraph(graph, { capacities });
  const byID = new Map(graph.units.map((unit) => [unit.unit_id, unit]));
  const ranks = criticalPathRanks(graph);
  const state = new Map(graph.units.map((unit) => [unit.unit_id, "pending"]));
  const queuedAt = new Map(graph.units.map((unit) => [unit.unit_id, 0]));
  const ageCredits = new Map(graph.units.map((unit) => [unit.unit_id, 0]));
  const activeClaims = new Map();
  const activeSharedLocks = new Map();
  const activeExclusiveLocks = new Set();
  const running = new Map();
  const waitSignatures = new Map();
  const events = [];
  const admissions = [];
  let seq = 0;
  let now = 0;
  const emit = (event, unit, status, extra) => {
    events.push(schedulerEvent(++seq, now, event, unit, status, extra));
  };
  for (const unit of graph.units) emit("queued", unit, "pending");

  while ([...state.values()].some((value) => value === "pending" || value === "running")) {
    if (cancelAtMs !== undefined && cancelAtMs <= now) {
      now = cancelAtMs;
      for (const unitID of [...running.keys()].sort(compareASCII)) {
        const unit = byID.get(unitID);
        addClaims(activeClaims, unit.resource_claims, -1);
        addLocks(activeSharedLocks, activeExclusiveLocks, unit, -1);
        running.delete(unitID);
        state.set(unitID, "cancelled");
        emit("cancelled", unit, "cancelled");
      }
      for (const unit of graph.units.filter((entry) => state.get(entry.unit_id) === "pending")) {
        state.set(unit.unit_id, "cancelled");
        emit("cancelled", unit, "cancelled");
      }
      break;
    }

    let skipped = false;
    do {
      skipped = false;
      for (const unit of graph.units) {
        if (state.get(unit.unit_id) !== "pending") continue;
        const dependency = failedDependency(unit, state, byID);
        if (!dependency) continue;
        state.set(unit.unit_id, "skipped");
        emit("skipped", unit, "skipped", { failure_reason: "dependency_failure" });
        skipped = true;
      }
    } while (skipped);

    let admitted = false;
    while (true) {
      const readyUnits = graph.units
        .filter((unit) => state.get(unit.unit_id) === "pending")
        .filter((unit) => ready(unit, state, byID));
      const candidates = readyUnits
        .filter((unit) => fitting(unit, capacities, activeClaims, activeSharedLocks, activeExclusiveLocks))
        .filter((unit) => !blockedByReadyExclusiveWaiter(unit, readyUnits))
        .sort((left, right) => {
          const leftAge = Math.floor((now - queuedAt.get(left.unit_id)) / agingQuantumMs);
          const rightAge = Math.floor((now - queuedAt.get(right.unit_id)) / agingQuantumMs);
          const leftRank =
            ranks.get(left.unit_id) +
            (leftAge + ageCredits.get(left.unit_id)) * agingQuantumMs;
          const rightRank =
            ranks.get(right.unit_id) +
            (rightAge + ageCredits.get(right.unit_id)) * agingQuantumMs;
          return rightRank - leftRank || compareASCII(left.unit_id, right.unit_id);
        });
      const unit = candidates[0];
      if (!unit) break;
      for (const waiting of readyUnits) {
        if (waiting.unit_id !== unit.unit_id) {
          ageCredits.set(waiting.unit_id, ageCredits.get(waiting.unit_id) + 1);
        }
      }
      admitted = true;
      state.set(unit.unit_id, "running");
      addClaims(activeClaims, unit.resource_claims, 1);
      addLocks(activeSharedLocks, activeExclusiveLocks, unit, 1);
      const duration = durations.get(unit.unit_id) ?? unit.estimated_work_ms;
      running.set(unit.unit_id, { finish: now + duration });
      admissions.push(unit.unit_id);
      emit("admitted", unit, "running");
      emit("started", unit, "running");
    }

    if (running.size === 0) {
      const pending = graph.units.filter((unit) => state.get(unit.unit_id) === "pending");
      if (pending.length === 0) break;
      throw new Error(
        `scheduler deadlock: ${pending.map((unit) => unit.unit_id).join(", ")}`,
      );
    }

    const nextFinish = Math.min(...[...running.values()].map((entry) => entry.finish));
    if (cancelAtMs !== undefined && cancelAtMs < nextFinish) {
      now = cancelAtMs;
      continue;
    }
    now = nextFinish;
    const finished = [...running.entries()]
      .filter(([, entry]) => entry.finish === nextFinish)
      .map(([unitID]) => unitID)
      .sort(compareASCII);
    for (const unitID of finished) {
      const unit = byID.get(unitID);
      running.delete(unitID);
      addClaims(activeClaims, unit.resource_claims, -1);
      addLocks(activeSharedLocks, activeExclusiveLocks, unit, -1);
      const outcome = outcomes.get(unitID) ?? "passed";
      if (outcome === "passed") {
        state.set(unitID, "passed");
        emit("completed", unit, "passed");
      } else {
        state.set(unitID, "failed");
        emit("failed", unit, "failed", { failure_class: outcome });
      }
    }
    if (!admitted && finished.length === 0) throw new Error("scheduler made no progress");
  }

  const failed = [...state.values()].some((value) =>
    ["failed", "cancelled"].includes(value),
  );
  return {
    status: failed ? "failed" : "passed",
    duration_ms: now,
    admissions,
    states: Object.fromEntries([...state.entries()].sort(([left], [right]) => compareASCII(left, right))),
    events,
  };
}

export async function runWorkGraph({
  graph,
  capacities,
  cwd,
  environment,
  executeUnit = executeUnitProcess,
  fixtureBroker,
  cache,
  signal,
  agingQuantumMs = 1000,
  cleanup = async () => {},
  onEvent = () => {},
}) {
  validateWorkGraph(graph, { capacities });
  cache?.validateGraph(graph);
  const byID = new Map(graph.units.map((unit) => [unit.unit_id, unit]));
  const ranks = criticalPathRanks(graph);
  const state = new Map(graph.units.map((unit) => [unit.unit_id, "pending"]));
  const queuedAt = new Map(graph.units.map((unit) => [unit.unit_id, 0]));
  const ageCredits = new Map(graph.units.map((unit) => [unit.unit_id, 0]));
  const activeClaims = new Map();
  const activeSharedLocks = new Map();
  const activeExclusiveLocks = new Set();
  const running = new Map();
  const waitSignatures = new Map();
  const events = [];
  const admissions = [];
  const terminalResults = new Map();
  const warmAffinities = new Set();
  const controller = new AbortController();
  const onAbort = () => controller.abort(signal?.reason);
  signal?.addEventListener("abort", onAbort, { once: true });
  if (signal?.aborted) onAbort();
  const started = performance.now();
  let seq = 0;
  const elapsed = () => Math.max(0, Math.floor(performance.now() - started));
  const emit = (event, unit, status, extra) => {
    const record = schedulerEvent(++seq, elapsed(), event, unit, status, extra);
    events.push(record);
    onEvent(record);
  };
  const runLifecycleUnit = {
    unit_id: "harness:run",
    resource_claims: {},
  };
  emit("run_started", runLifecycleUnit, "running");
  for (const unit of graph.units) emit("queued", unit, "pending");

  let cleanupError = null;

  try {
    while ([...state.values()].some((value) => value === "pending" || value === "running")) {
      let skipped = false;
      do {
        skipped = false;
        for (const unit of graph.units) {
          if (state.get(unit.unit_id) !== "pending") continue;
          const dependency = failedDependency(unit, state, byID);
          if (!dependency) continue;
          state.set(unit.unit_id, "skipped");
          terminalResults.set(unit.unit_id, {
            status: "skipped",
            failure_reason: "dependency_failure",
          });
          emit("skipped", unit, "skipped", { failure_reason: "dependency_failure" });
          skipped = true;
        }
      } while (skipped);

      if (!controller.signal.aborted) {
        while (true) {
          const now = elapsed();
          const readyUnits = graph.units
            .filter((entry) => state.get(entry.unit_id) === "pending")
            .filter((entry) => ready(entry, state, byID));
          for (const waiting of readyUnits.filter(
            (entry) => !fitting(entry, capacities, activeClaims, activeSharedLocks, activeExclusiveLocks),
          )) {
            const details = blockingDetails(
              waiting,
              capacities,
              activeClaims,
              activeSharedLocks,
              activeExclusiveLocks,
              running,
              byID,
            );
            const signature = JSON.stringify(details);
            if (waitSignatures.get(waiting.unit_id) !== signature) {
              waitSignatures.set(waiting.unit_id, signature);
              emit("resource_wait", waiting, "pending", details);
            }
          }
          const unit = readyUnits
            .filter((entry) => fitting(entry, capacities, activeClaims, activeSharedLocks, activeExclusiveLocks))
            .filter((entry) => !blockedByReadyExclusiveWaiter(entry, readyUnits))
            .sort((left, right) => {
              const leftWarm = warmAffinities.has(left.affinity_key) ? 1 : 0;
              const rightWarm = warmAffinities.has(right.affinity_key) ? 1 : 0;
              if (leftWarm !== rightWarm) return rightWarm - leftWarm;
              const leftAge = Math.floor((now - queuedAt.get(left.unit_id)) / agingQuantumMs);
              const rightAge = Math.floor((now - queuedAt.get(right.unit_id)) / agingQuantumMs);
              const leftRank =
                ranks.get(left.unit_id) +
                (leftAge + ageCredits.get(left.unit_id)) * agingQuantumMs;
              const rightRank =
                ranks.get(right.unit_id) +
                (rightAge + ageCredits.get(right.unit_id)) * agingQuantumMs;
              return rightRank - leftRank || compareASCII(left.unit_id, right.unit_id);
            })[0];
          if (!unit) break;
          for (const waiting of readyUnits) {
            if (waiting.unit_id !== unit.unit_id) {
              ageCredits.set(waiting.unit_id, ageCredits.get(waiting.unit_id) + 1);
            }
          }
          state.set(unit.unit_id, "running");
          if (unit.affinity_key) warmAffinities.delete(unit.affinity_key);
          waitSignatures.delete(unit.unit_id);
          addClaims(activeClaims, unit.resource_claims, 1);
          addLocks(activeSharedLocks, activeExclusiveLocks, unit, 1);
          admissions.push(unit.unit_id);
          emit("admitted", unit, "running");
          emit("started", unit, "running");
          const promise = Promise.resolve()
            .then(async () => {
              const cacheResult = cache
                ? await cache.lookup(unit)
                : { outcome: "bypass", reason: "cache_unconfigured" };
              emit(`cache_${cacheResult.outcome}`, unit, cacheResult.outcome === "hit" ? "passed" : "running", {
                ...(cacheResult.profile_id ? { cache_profile_id: cacheResult.profile_id } : {}),
                cache_reason: cacheResult.reason,
                ...(cacheResult.output_digest ? { output_digest: cacheResult.output_digest } : {}),
              });
              if (cacheResult.outcome === "hit") {
                return { status: "passed", cache: cacheResult };
              }
              const lease = fixtureBroker && unit.fixture_lease !== "none"
                ? await fixtureBroker.acquire(unit.fixture_lease, {
                    affinityKey: unit.affinity_key ?? unit.owner_id,
                    unitID: unit.unit_id,
                    runtimeProfileID:
                      unit.command.environment.CARTULARY_BROWSER_RUNTIME_PROFILE_ID,
                    fixtureProfileID:
                      unit.command.environment.CARTULARY_FIXTURE_PROFILE_ID,
                    snapshotKey:
                      unit.command.environment.CARTULARY_FIXTURE_SNAPSHOT_KEY,
                    builderUnitID:
                      unit.command.environment.CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID,
                    rowID: unit.command.environment.CARTULARY_FIXTURE_ROW_ID,
                    predicateID:
                      unit.command.environment.CARTULARY_FIXTURE_PREDICATE_ID,
                  })
                : null;
              if (lease) {
                emit("fixture_acquired", unit, "running", {
                  fixture_lease_id: lease.record.lease_id,
                });
              }
              let result;
              try {
                result = await executeUnit(unit, {
                  cwd,
                  environment,
                  fixtureLease: lease,
                  signal: controller.signal,
                });
                if (result.status === "passed") await cache?.store(unit);
              } finally {
                if (lease) {
                  const healthy =
                    !controller.signal.aborted &&
                    (unit.fixture_lease !== "browser_stack" || result?.status === "passed");
                  try {
                    const release = await lease.release({
                      healthy,
                      retainWarm:
                        unit.command.environment.CARTULARY_BROWSER_RELEASE_AFFINITY !== "1",
                    });
                    if (release.retained) result = { ...result, retained_fixture: true };
                    emit("fixture_released", unit, "running", {
                      fixture_lease_id: lease.record.lease_id,
                    });
                  } catch (error) {
                    emit("fixture_released", unit, "failed", {
                      fixture_lease_id: lease.record.lease_id,
                      failure_class: "harness",
                      failure_reason: "cleanup_error",
                    });
                    if (result && result.status !== "passed") {
                      result = { ...result, cleanup_error: error };
                    } else {
                      throw error;
                    }
                  }
                }
              }
              return result;
            })
            .then((result) => ({ unit_id: unit.unit_id, result }))
            .catch((error) => ({
              unit_id: unit.unit_id,
              result: {
                status: "failed",
                failure_class: "infra",
                failure_reason: "fixture_error",
                error,
              },
            }));
          running.set(unit.unit_id, promise);
        }
      }

      if (running.size === 0) {
        const pending = graph.units.filter((unit) => state.get(unit.unit_id) === "pending");
        if (pending.length === 0) break;
        if (controller.signal.aborted) {
          for (const unit of pending) {
            state.set(unit.unit_id, "cancelled");
            terminalResults.set(unit.unit_id, {
              status: "cancelled",
              failure_class: "interrupted",
              failure_reason: "cancelled_or_interrupted",
            });
            emit("cancelled", unit, "cancelled", {
              failure_class: "interrupted",
              failure_reason: "cancelled_or_interrupted",
            });
          }
          break;
        }
        throw new Error(`scheduler deadlock: ${pending.map((unit) => unit.unit_id).join(", ")}`);
      }

      const { unit_id: unitID, result } = await Promise.race(running.values());
      const unit = byID.get(unitID);
      running.delete(unitID);
      addClaims(activeClaims, unit.resource_claims, -1);
      addLocks(activeSharedLocks, activeExclusiveLocks, unit, -1);
      terminalResults.set(unitID, result);
      if (unit.affinity_key) {
        if (result.retained_fixture) warmAffinities.add(unit.affinity_key);
        else warmAffinities.delete(unit.affinity_key);
      }
      if (result.status === "passed") {
        state.set(unitID, "passed");
        emit("completed", unit, "passed");
      } else if (result.status === "cancelled" || controller.signal.aborted) {
        state.set(unitID, "cancelled");
        emit("cancelled", unit, "cancelled", {
          failure_class: "interrupted",
          failure_reason: "cancelled_or_interrupted",
        });
      } else {
        state.set(unitID, "failed");
        emit("failed", unit, "failed", {
          failure_class: result.failure_class ?? "harness",
          failure_reason: result.failure_reason ?? "execution_failure",
        });
      }
    }
  } finally {
    controller.abort();
    await Promise.allSettled(running.values());
    emit("cleanup_started", runLifecycleUnit, "running");
    const cleanupResults = await Promise.allSettled([
      fixtureBroker?.close(),
      cleanup(),
    ]);
    cleanupError = cleanupResults.find((result) => result.status === "rejected")?.reason ?? null;
    emit(
      "cleanup_completed",
      runLifecycleUnit,
      cleanupError ? "failed" : "passed",
      cleanupError
        ? { failure_class: "harness", failure_reason: "cleanup_error" }
        : undefined,
    );
    signal?.removeEventListener("abort", onAbort);
  }
  const cancelled = [...state.values()].some((value) => value === "cancelled");
  const failed = cleanupError || [...state.values()].some((value) =>
    ["failed", "cancelled"].includes(value),
  );
  emit(
    "run_completed",
    runLifecycleUnit,
    cancelled ? "cancelled" : failed ? "failed" : "passed",
    cleanupError
      ? { failure_class: "harness", failure_reason: "cleanup_error" }
      : undefined,
  );
  return {
    status: failed ? "failed" : "passed",
    duration_ms: events.at(-1).monotonic_ms,
    cleanup_error: cleanupError ? String(cleanupError.message ?? cleanupError) : null,
    admissions,
    states: Object.fromEntries([...state.entries()].sort(([left], [right]) => compareASCII(left, right))),
    unit_results: Object.fromEntries(
      [...terminalResults.entries()].sort(([left], [right]) => compareASCII(left, right)),
    ),
    events,
  };
}
