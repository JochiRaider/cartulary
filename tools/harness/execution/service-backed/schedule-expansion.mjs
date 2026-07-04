import {
  browserGroupCompletionKey,
  browserGroupNeeds,
  browserGroupWorkerEnvFromPlan,
  browserGroupWorkerSlotPlan,
  browserStageCompletionNeeds,
  browserStageSessionKey,
} from "../../scheduler/adapters/browser.mjs";
import { collectGoShardsForTarget } from "../../backend/backend-shard-plan.mjs";

const serviceSessionResource = "suite_service_stack";
const goCPUResource = "go_cpu";
const goIOResource = "go_io";
const hostCPUResource = "host_cpu";
const hostIOResource = "host_io";
const postgresResetResource = "postgres_reset";
const postgresCloneResource = "postgres_clone";
const buildServerTarget = "build-server";
const buildMigrateTarget = "build-migrate";
const buildWebTarget = "build-web";
const testServiceImagesTarget = "test-service-images";
const defaultSchedulerPriority = 0;

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
}

function resourceClaimsObject(value) {
  return Object.fromEntries(
    Object.entries(value).sort(([left], [right]) => left.localeCompare(right)),
  );
}

function addClaim(claims, resource, amount) {
  if (amount === "limit") {
    claims.set(resource, amount);
    return;
  }
  if (!Number.isInteger(amount) || amount < 1) {
    throw new Error(`resource claim ${resource} must be a positive integer or "limit"`);
  }
  if (claims.get(resource) === "limit") {
    return;
  }
  claims.set(resource, (claims.get(resource) ?? 0) + amount);
}

export function mapServiceBackedClaimsToCheckClaims(rawClaims, { ensureHost = false } = {}) {
  const claims = new Map();
  for (const [resource, amount] of Object.entries(requireObject(rawClaims, "resource_claims"))) {
    if (resource === goCPUResource) {
      addClaim(claims, hostCPUResource, amount);
    } else if (resource === goIOResource) {
      addClaim(claims, hostIOResource, amount);
    } else {
      addClaim(claims, resource, amount);
    }
  }
  if (ensureHost) {
    if (!claims.has(hostCPUResource)) {
      claims.set(hostCPUResource, 1);
    }
    if (!claims.has(hostIOResource)) {
      claims.set(hostIOResource, 1);
    }
  }
  return resourceClaimsObject(Object.fromEntries(claims.entries()));
}

function checkClaimsForShard(source, shard) {
  const claims = new Map(Object.entries(mapServiceBackedClaimsToCheckClaims(source.resource_claims)));
  switch (shard.scheduler_profile) {
    case "cpu_heavy":
      addClaim(claims, hostCPUResource, 2);
      addClaim(claims, hostIOResource, 1);
      break;
    case "io_heavy":
      addClaim(claims, hostCPUResource, 1);
      addClaim(claims, hostIOResource, 2);
      break;
    case "reset_heavy":
      addClaim(claims, hostCPUResource, 1);
      addClaim(claims, hostIOResource, 2);
      addClaim(claims, postgresResetResource, 1);
      break;
    case "clone_heavy":
      addClaim(claims, hostCPUResource, 1);
      addClaim(claims, hostIOResource, 2);
      addClaim(claims, postgresCloneResource, 1);
      break;
    case "transaction_heavy":
      addClaim(claims, hostCPUResource, 1);
      addClaim(claims, hostIOResource, 1);
      break;
    default:
      addClaim(claims, hostCPUResource, 1);
      addClaim(claims, hostIOResource, 1);
      break;
  }
  return resourceClaimsObject(Object.fromEntries(claims.entries()));
}

function shardCompletionKey(shardName) {
  return `go_shard:${shardName}`;
}

function sourceNeeds(source, serviceSessionKey, extraNeeds = []) {
  return [serviceSessionKey, ...extraNeeds, ...(source.needs ?? [])];
}

function sortedUnique(values) {
  return Array.from(new Set(values.filter((value) => value !== ""))).sort((left, right) =>
    String(left).localeCompare(String(right)),
  );
}

function runtimeBinaryIDsForShard(shard) {
  return sortedUnique((shard.items ?? []).flatMap((item) => item.runtime_binaries ?? []));
}

function runtimeBinaryRegistry(source) {
  return new Map((source.runtime_binary_records ?? []).map((entry) => [entry.id, entry]));
}

function runtimeBinaryEnvForIDs(source, ids) {
  const registry = runtimeBinaryRegistry(source);
  const env = {};
  for (const id of ids) {
    const entry = registry.get(id);
    if (!entry) {
      throw new Error(`${source.target} shard runtime binary ${id} is missing from runtime_binary_records`);
    }
    if (id !== "operator") {
      throw new Error(`${source.target} shard runtime binary ${id} is missing default output path wiring`);
    }
    env[entry.consumer_env] = "operator";
  }
  return env;
}

function runtimeBinaryNeedsForIDs(source, ids) {
  const registry = runtimeBinaryRegistry(source);
  return ids.map((id) => {
    const entry = registry.get(id);
    if (!entry) {
      throw new Error(`${source.target} shard runtime binary ${id} is missing from runtime_binary_records`);
    }
    return entry.producer_target;
  });
}

function mergeEnv(...parts) {
  const entries = new Map();
  for (const part of parts) {
    for (const [name, value] of Object.entries(part ?? {})) {
      entries.set(name, value);
    }
  }
  return Object.fromEntries([...entries.entries()].sort(([left], [right]) => left.localeCompare(right)));
}

function shardRuntimeConfig(source, shard) {
  const runtimeBinaries = runtimeBinaryIDsForShard(shard);
  return {
    runtimeBinaries,
    needs: runtimeBinaryNeedsForIDs(source, runtimeBinaries),
    env: runtimeBinaryEnvForIDs(source, runtimeBinaries),
  };
}

function serviceSessionNeeds(parentNeeds) {
  return parentNeeds.filter((need) => need === testServiceImagesTarget);
}

function browserStageExtraNeeds(parentNeeds) {
  return [buildWebTarget, buildServerTarget, buildMigrateTarget].filter((need) =>
    parentNeeds.includes(need),
  );
}

function isRetainedBrowserStageResource(resource) {
  return resource === "browser_stack" || resource === "process" || resource.startsWith("browser_stage_");
}

function retainedBrowserStageClaimsFromEntries(entries) {
  return resourceClaimsObject(
    Object.fromEntries(
      entries.filter(([resource]) => isRetainedBrowserStageResource(resource)),
    ),
  );
}

function retainedBrowserStageClaims(rawClaims) {
  const mapped = mapServiceBackedClaimsToCheckClaims(rawClaims, { ensureHost: true });
  return retainedBrowserStageClaimsFromEntries(Object.entries(mapped));
}

function browserGroupClaims(rawClaims) {
  return mapServiceBackedClaimsToCheckClaims(rawClaims ?? {}, { ensureHost: true });
}

function browserSessionGroupName(source) {
  const raw = source.browser_session_group ?? source.browserSessionGroup;
  if (typeof raw === "string" && raw.trim() !== "") {
    return raw.trim();
  }
  return source.target;
}

function browserSessionIsolationReason(source) {
  const raw =
    source.browser_session_isolation_reason ?? source.browserSessionIsolationReason;
  return typeof raw === "string" && raw.trim() !== "" ? raw.trim() : "";
}

function browserSessionRetainsOwnTarget(source) {
  return browserSessionGroupName(source) === source.target;
}

function browserSessionFinalizerTarget(sessionGroup) {
  return `browser-session-${sessionGroup.replaceAll(/[^A-Za-z0-9_.-]/g, "-")}`;
}

function browserSessionFinalizerCompletionKey(sessionGroup) {
  return `browser_session_finalizer:${sessionGroup}`;
}

function sharedBrowserSession(info) {
  return (info?.sources?.length ?? 0) > 1;
}

function browserSessionWorkUnitID(scheduleTarget, source) {
  const group = browserSessionGroupName(source);
  if (browserSessionRetainsOwnTarget(source)) {
    return `${scheduleTarget}:browser-stage-session:${source.browser_stage}`;
  }
  return `${scheduleTarget}:browser-stage-session:${group}`;
}

function mergeSessionClaims(left, right) {
  const claims = new Map(Object.entries(left ?? {}));
  for (const [resource, amount] of Object.entries(right ?? {})) {
    const previous = claims.get(resource);
    if (amount === "limit" || previous === "limit") {
      claims.set(resource, "limit");
      continue;
    }
    if (!Number.isInteger(amount) || amount < 1) {
      throw new Error(`resource claim ${resource} must be a positive integer or "limit"`);
    }
    claims.set(resource, Math.max(Number.isInteger(previous) ? previous : 0, amount));
  }
  return resourceClaimsObject(Object.fromEntries(claims.entries()));
}

function unionStrings(left, right) {
  return Array.from(new Set([...(left ?? []), ...(right ?? [])]));
}

function browserSessionInfos(sources, { serviceSessionKey, parentNeeds }) {
  const infos = new Map();
  for (const [sourceIndex, source] of sources.entries()) {
    if (source.type !== "browser_stage") {
      continue;
    }
    const group = browserSessionGroupName(source);
    const sourceClaims = mapServiceBackedClaimsToCheckClaims(source.resource_claims, {
      ensureHost: true,
    });
    const retainedClaims = retainedBrowserStageClaims(source.resource_claims);
    const needs = sourceNeeds(source, serviceSessionKey, browserStageExtraNeeds(parentNeeds));
    const groupNeeds = browserStageCompletionNeeds(source.groups);
    const info = infos.get(group) ?? {
      group,
      sessionKey: browserStageSessionKey(group),
      firstSource: source,
      firstSourceIndex: sourceIndex,
      finalizerSource: source,
      finalizerIndex: sourceIndex,
      resourceClaims: {},
      retainedClaims: {},
      needs: [],
      groupNeeds: [],
      priority: 0,
      weightMs: 0,
      isolationReason: "",
      sources: [],
    };
    info.sources.push(source);
    info.resourceClaims = mergeSessionClaims(info.resourceClaims, sourceClaims);
    info.retainedClaims = mergeSessionClaims(info.retainedClaims, retainedClaims);
    info.needs = unionStrings(info.needs, needs);
    info.groupNeeds = unionStrings(info.groupNeeds, groupNeeds);
    info.priority = Math.max(info.priority, priority(source.priority));
    info.weightMs = Math.max(info.weightMs, source.weight_ms);
    const isolationReason = browserSessionIsolationReason(source);
    if (isolationReason) {
      info.isolationReason = isolationReason;
    }
    infos.set(group, info);
  }
  return infos;
}

function priority(value) {
  return Number.isInteger(value) && value > 0 ? value : defaultSchedulerPriority;
}

function command(type, extra = {}) {
  return { type, ...extra };
}

export function expandServiceBackedScheduleForCheck({
  repoRoot,
  serviceSchedule,
  parentUnit,
}) {
  const scheduleTarget = serviceSchedule.target;
  const serviceSessionKey = `service_session:${scheduleTarget}`;
  const serviceWeightMs = parentUnit.weight_ms;
  const parentNeeds = [...(parentUnit.needs ?? [])];
  const serviceNeeds = serviceSessionNeeds(parentNeeds);
  const serviceCompletePriority = priority(serviceSchedule.service_complete_priority);
  const sessionInfos = browserSessionInfos(serviceSchedule.work_unit_sources ?? [], {
    serviceSessionKey,
    parentNeeds,
  });
  const browserWorkerSlotPlan = browserGroupWorkerSlotPlan(serviceSchedule.work_unit_sources ?? []);
  const expanded = [
    {
      id: `${scheduleTarget}:service-session`,
      kind: "service_session",
      target: scheduleTarget,
      label: `${scheduleTarget}/service-session`,
      priority: priority(parentUnit.priority),
      weight_ms: serviceWeightMs,
      needs: serviceNeeds,
      completion_keys: [serviceSessionKey],
      resource_claims: {
        host_cpu: 1,
        host_io: 1,
        [serviceSessionResource]: 1,
      },
      retained_resource_claims: {
        [serviceSessionResource]: 1,
      },
      service_session: {
        target: scheduleTarget,
      },
      command: command("service_session_start", { service_target: scheduleTarget }),
    },
  ];

  for (const [sourceIndex, source] of (serviceSchedule.work_unit_sources ?? []).entries()) {
    if (source.type === "browser_stage") {
      const sessionGroup = browserSessionGroupName(source);
      const sessionInfo = sessionInfos.get(sessionGroup);
      const isSessionOwner = sessionInfo?.firstSource === source;
      const isSessionFinalizer = sessionInfo?.finalizerSource === source;
      const finalizeOnStageComplete = isSessionFinalizer && !sharedBrowserSession(sessionInfo);
      if (isSessionOwner) {
        expanded.push({
          id: browserSessionWorkUnitID(scheduleTarget, source),
          kind: "browser_stage_session",
          target: source.target,
          label: `${sessionGroup}/stage-session`,
          aggregate_target: source.target,
          priority: sessionInfo.priority,
          weight_ms: sessionInfo.weightMs,
          needs: sessionInfo.needs,
          completion_keys: [sessionInfo.sessionKey],
          failure_keys: [sessionInfo.sessionKey],
          resource_claims: sessionInfo.resourceClaims,
          retained_resource_claims: sessionInfo.retainedClaims,
          service_session: {
            target: scheduleTarget,
          },
          browser_stage: source.browser_stage,
          browser_session_group: sessionGroup,
          ...(sessionInfo.isolationReason
            ? { browser_session_isolation_reason: sessionInfo.isolationReason }
            : {}),
          command: command("browser_stage_session_start", {
            service_target: scheduleTarget,
            browser_stage: source.browser_stage,
          }),
          order: sourceIndex,
        });
      }
      expanded.push({
        id: `${scheduleTarget}:browser-stage-complete:${source.browser_stage}`,
        kind: "browser_stage_complete",
        target: source.target,
        label: `${source.target}/complete`,
        aggregate_target: source.target,
        priority: priority(source.priority),
        weight_ms: 1,
        needs: browserStageCompletionNeeds(source.groups),
        completion_keys: [source.target],
        failure_keys: [source.target],
        count_in_total: false,
        counts_started: false,
        resource_claims: {},
        release_retained_resource_claims: finalizeOnStageComplete
          ? (sessionInfo?.retainedClaims ?? {})
          : {},
        service_session: {
          target: scheduleTarget,
        },
        browser_stage: source.browser_stage,
        browser_session_group: sessionGroup,
        browser_session_finalizer: finalizeOnStageComplete,
        ...(sessionInfo?.isolationReason
          ? { browser_session_isolation_reason: sessionInfo.isolationReason }
          : {}),
        command: command("browser_stage_complete", {
          service_target: scheduleTarget,
          browser_stage: source.browser_stage,
        }),
        order: sourceIndex,
      });
      for (const group of source.groups ?? []) {
        expanded.push({
          id: `${scheduleTarget}:${group.id}`,
          kind: "browser_group",
          target: group.target,
          label: `${source.target}/${group.name}`,
          aggregate_target: source.target,
          priority: priority(group.priority ?? source.priority),
          weight_ms: group.weight_ms,
          needs: browserGroupNeeds(sessionInfo?.sessionKey ?? browserStageSessionKey(source.target)),
          completion_keys: [browserGroupCompletionKey(group.id)],
          failure_keys: [browserGroupCompletionKey(group.id)],
          resource_claims: browserGroupClaims(group.resource_claims),
          service_session: {
            target: scheduleTarget,
          },
          browser_stage: source.browser_stage,
          browser_session_group: sessionGroup,
          browser_group: clone(group),
          env: browserGroupWorkerEnvFromPlan(browserWorkerSlotPlan, group),
          command: command("browser_group", {
            service_target: scheduleTarget,
            browser_stage: source.browser_stage,
            group_id: group.id,
          }),
          order: sourceIndex,
        });
      }
      continue;
    }

    if (source.type === "make_target") {
      expanded.push({
        id: `${scheduleTarget}:${source.target}`,
        kind: "service_make_target",
        target: source.target,
        label: source.target,
        aggregate_target: source.target,
        priority: priority(source.priority),
        weight_ms: source.weight_ms,
        needs: sourceNeeds(source, serviceSessionKey),
        ...(source.env ? { env: clone(source.env) } : {}),
        ...(source.runtime_binaries ? { runtime_binaries: clone(source.runtime_binaries) } : {}),
        completion_keys: [source.target],
        failure_keys: [source.target],
        resource_claims: mapServiceBackedClaimsToCheckClaims(source.resource_claims, {
          ensureHost: true,
        }),
        service_session: {
          target: scheduleTarget,
        },
        command: command("make_target", { target: source.target, service_target: scheduleTarget }),
        order: sourceIndex,
      });
      continue;
    }

    if (source.type !== "go_shards") {
      throw new Error(`${scheduleTarget} source ${source.target} has unsupported type ${source.type}`);
    }

    const shards = collectGoShardsForTarget(repoRoot, source.target, {
      defaultCheckOnly: source.default_check_required === true,
    });
    if (shards.length === 0) {
      throw new Error(`${scheduleTarget} go_shards source ${source.target} selected no shards`);
    }
    expanded.push({
      id: `${scheduleTarget}:finalize:${source.target}`,
      kind: "aggregate_finalize",
      target: source.target,
      label: `finalize/${source.target}`,
      aggregate_target: source.target,
      priority: priority(source.priority),
      weight_ms: 1,
      needs: shards.map((shard) => shardCompletionKey(shard.name)),
      completion_keys: [source.target],
      failure_keys: [source.target],
      count_in_total: false,
      counts_started: false,
      resource_claims: {},
      service_session: {
        target: scheduleTarget,
      },
      shard_names: shards.map((shard) => shard.name),
      unblock_label: source.target,
      command: command("go_shard_finalize", {
        target: source.target,
        service_target: scheduleTarget,
      }),
      order: sourceIndex,
    });
    for (const shard of shards) {
      const runtime = shardRuntimeConfig(source, shard);
      const env = mergeEnv(source.env, runtime.env);
      expanded.push({
        id: `${scheduleTarget}:${source.target}:${shard.name}`,
        kind: "go_shard",
        target: source.target,
        label: `${source.target}/${shard.name}`,
        aggregate_target: source.target,
        priority: priority(source.priority),
        weight_ms: shard.weight_ms,
        needs: sourceNeeds(source, serviceSessionKey, runtime.needs),
        ...(Object.keys(env).length > 0 ? { env } : {}),
        ...(runtime.runtimeBinaries.length > 0 ? { runtime_binaries: runtime.runtimeBinaries } : {}),
        completion_keys: [shardCompletionKey(shard.name)],
        failure_keys: [shardCompletionKey(shard.name)],
        running_dependency_keys: [source.target],
        complete_on_failure: true,
        shard: shard.name,
        scheduler_profile: shard.scheduler_profile,
        resource_claims: checkClaimsForShard(source, shard),
        service_session: {
          target: scheduleTarget,
        },
        command: command("go_shard", {
          target: source.target,
          shard: shard.name,
          service_target: scheduleTarget,
        }),
        order: sourceIndex,
      });
    }
  }

  const sharedSessionFinalizerKeys = [];
  for (const info of sessionInfos.values()) {
    if (!sharedBrowserSession(info)) {
      continue;
    }
    const finalizerKey = browserSessionFinalizerCompletionKey(info.group);
    sharedSessionFinalizerKeys.push(finalizerKey);
    expanded.push({
      id: `${scheduleTarget}:browser-session-finalizer:${info.group}`,
      kind: "browser_session_finalizer",
      target: browserSessionFinalizerTarget(info.group),
      label: `${info.group}/session-finalizer`,
      aggregate_target: browserSessionFinalizerTarget(info.group),
      priority: info.priority,
      weight_ms: 1,
      needs: info.groupNeeds,
      completion_keys: [finalizerKey],
      failure_keys: [finalizerKey],
      count_in_total: false,
      counts_started: false,
      resource_claims: {},
      release_retained_resource_claims: info.retainedClaims,
      service_session: {
        target: scheduleTarget,
      },
      browser_session_group: info.group,
      ...(info.isolationReason
        ? { browser_session_isolation_reason: info.isolationReason }
        : {}),
      command: command("browser_session_finalizer", {
        service_target: scheduleTarget,
        browser_session_group: info.group,
      }),
      order: info.finalizerIndex,
    });
  }

  expanded.push({
    id: `${scheduleTarget}:complete`,
    kind: "service_complete",
    target: scheduleTarget,
    label: `${scheduleTarget}/complete`,
    priority: serviceCompletePriority,
    weight_ms: 1,
    needs: [
      ...(serviceSchedule.work_unit_sources ?? []).map((source) => source.target),
      ...sharedSessionFinalizerKeys,
    ],
    completion_keys: [scheduleTarget],
    failure_keys: [scheduleTarget],
    produces_summary_targets: parentUnit.produces_summary_targets ?? [scheduleTarget],
    count_in_total: false,
    counts_started: false,
    resource_claims: {},
    service_session: {
      target: scheduleTarget,
    },
    command: command("service_complete", { service_target: scheduleTarget }),
  });

  return expanded
    .sort(
      (left, right) =>
        (right.priority ?? 0) - (left.priority ?? 0) ||
        right.weight_ms - left.weight_ms ||
        (left.order ?? 0) - (right.order ?? 0) ||
        left.id.localeCompare(right.id),
    )
    .map(({ order: _order, ...unit }) => clone(unit));
}

function schedulerClaimsForShard(shard) {
  switch (shard.scheduler_profile) {
    case "cpu_heavy":
      return {
        [goCPUResource]: 2,
        [goIOResource]: 1,
      };
    case "io_heavy":
      return {
        [goCPUResource]: 1,
        [goIOResource]: 2,
      };
    case "reset_heavy":
      return {
        [goCPUResource]: 1,
        [goIOResource]: 2,
        [postgresResetResource]: 1,
      };
    case "clone_heavy":
      return {
        [goCPUResource]: 1,
        [goIOResource]: 2,
        [postgresCloneResource]: 1,
      };
    case "transaction_heavy":
      return {
        [goCPUResource]: 1,
        [goIOResource]: 1,
      };
    default:
      return {
        [goCPUResource]: 1,
        [goIOResource]: 1,
      };
  }
}

function mergeClaims(left, ...claimObjects) {
  const claims = new Map(Object.entries(left ?? {}));
  for (const claimObject of claimObjects) {
    for (const [resource, amount] of Object.entries(claimObject ?? {})) {
      addClaim(claims, resource, amount);
    }
  }
  return resourceClaimsObject(Object.fromEntries(claims.entries()));
}

function directRetainedBrowserStageClaims(rawClaims) {
  return retainedBrowserStageClaimsFromEntries(Object.entries(rawClaims ?? {}));
}

function directRuntimeProducerClaims() {
  return {
    [goCPUResource]: 1,
    [goIOResource]: 1,
  };
}

function addDirectRuntimeProducerUnits(unitsByTarget, runtime, source, sourceIndex) {
  for (const target of runtime.needs) {
    if (unitsByTarget.has(target)) {
      continue;
    }
    unitsByTarget.set(target, {
      id: target,
      kind: "make_target",
      class: "backend",
      target,
      label: target,
      aggregate_target: target,
      priority: priority(source.priority),
      weight_ms: 1,
      needs: [],
      completion_keys: [target],
      failure_keys: [target],
      resource_claims: directRuntimeProducerClaims(),
      command: command("make_target", { target }),
      order: sourceIndex - 0.5,
    });
  }
}

function directBrowserSessionInfos(sources) {
  const infos = new Map();
  for (const [sourceIndex, source] of sources.entries()) {
    if (source.type !== "browser_stage") {
      continue;
    }
    const group = browserSessionGroupName(source);
    const info = infos.get(group) ?? {
      group,
      sessionKey: browserStageSessionKey(group),
      firstSource: source,
      firstSourceIndex: sourceIndex,
      finalizerSource: source,
      finalizerIndex: sourceIndex,
      resourceClaims: {},
      retainedClaims: {},
      needs: [],
      groupNeeds: [],
      priority: 0,
      weightMs: 0,
      isolationReason: "",
      sources: [],
    };
    info.sources.push(source);
    info.resourceClaims = mergeSessionClaims(
      info.resourceClaims,
      resourceClaimsObject(source.resource_claims ?? {}),
    );
    info.retainedClaims = mergeSessionClaims(
      info.retainedClaims,
      directRetainedBrowserStageClaims(source.resource_claims),
    );
    info.needs = unionStrings(info.needs, source.needs ?? []);
    info.groupNeeds = unionStrings(info.groupNeeds, browserStageCompletionNeeds(source.groups));
    info.priority = Math.max(info.priority, priority(source.priority));
    info.weightMs = Math.max(info.weightMs, source.weight_ms);
    const isolationReason = browserSessionIsolationReason(source);
    if (isolationReason) {
      info.isolationReason = isolationReason;
    }
    infos.set(group, info);
  }
  return infos;
}

export function expandServiceBackedSchedule({
  repoRoot,
  serviceSchedule,
}) {
  const scheduleTarget = serviceSchedule.target;
  const counted = [];
  const aggregate = [];
  const shardWorkByName = new Map();
  const runtimeProducerUnitsByTarget = new Map();
  const sessionInfos = directBrowserSessionInfos(serviceSchedule.work_unit_sources ?? []);
  const browserWorkerSlotPlan = browserGroupWorkerSlotPlan(serviceSchedule.work_unit_sources ?? []);

  for (const [sourceIndex, source] of (serviceSchedule.work_unit_sources ?? []).entries()) {
    if (source.type === "browser_stage") {
      const sessionGroup = browserSessionGroupName(source);
      const sessionInfo = sessionInfos.get(sessionGroup);
      const isSessionOwner = sessionInfo?.firstSource === source;
      const isSessionFinalizer = sessionInfo?.finalizerSource === source;
      const finalizeOnStageComplete = isSessionFinalizer && !sharedBrowserSession(sessionInfo);
      if (isSessionOwner) {
        counted.push({
          id: browserSessionRetainsOwnTarget(source)
            ? `browser-stage-session:${source.browser_stage}`
            : `browser-stage-session:${sessionGroup}`,
          kind: "browser_stage_session",
          class: source.class,
          target: source.target,
          label: `${sessionGroup}/stage-session`,
          aggregate_target: source.target,
          priority: sessionInfo.priority,
          weight_ms: sessionInfo.weightMs,
          needs: sessionInfo.needs,
          completion_keys: [sessionInfo.sessionKey],
          failure_keys: [sessionInfo.sessionKey],
          resource_claims: sessionInfo.resourceClaims,
          retained_resource_claims: sessionInfo.retainedClaims,
          browser_stage: source.browser_stage,
          browser_session_group: sessionGroup,
          ...(sessionInfo.isolationReason
            ? { browser_session_isolation_reason: sessionInfo.isolationReason }
            : {}),
          command: command("browser_stage_session_start", {
            service_target: scheduleTarget,
            browser_stage: source.browser_stage,
          }),
          order: sourceIndex,
        });
      }
      aggregate.push({
        id: `browser-stage-complete:${source.browser_stage}`,
        kind: "browser_stage_complete",
        class: source.class,
        target: source.target,
        label: `${source.target}/complete`,
        aggregate_target: source.target,
        priority: priority(source.priority),
        weight_ms: 1,
        needs: browserStageCompletionNeeds(source.groups),
        completion_keys: [source.target],
        failure_keys: [source.target],
        count_in_total: false,
        counts_started: false,
        resource_claims: {},
        release_retained_resource_claims: finalizeOnStageComplete
          ? (sessionInfo?.retainedClaims ?? {})
          : {},
        browser_stage: source.browser_stage,
        browser_session_group: sessionGroup,
        browser_session_finalizer: finalizeOnStageComplete,
        ...(sessionInfo?.isolationReason
          ? { browser_session_isolation_reason: sessionInfo.isolationReason }
          : {}),
        command: command("browser_stage_complete", {
          service_target: scheduleTarget,
          browser_stage: source.browser_stage,
        }),
        order: sourceIndex,
      });
      for (const group of source.groups ?? []) {
        counted.push({
          id: group.id,
          kind: "browser_group",
          class: source.class,
          target: group.target,
          label: `${source.target}/${group.name}`,
          aggregate_target: source.target,
          priority: priority(group.priority ?? source.priority),
          weight_ms: group.weight_ms,
          needs: browserGroupNeeds(sessionInfo?.sessionKey ?? browserStageSessionKey(source.target)),
          completion_keys: [browserGroupCompletionKey(group.id)],
          failure_keys: [browserGroupCompletionKey(group.id)],
          resource_claims: resourceClaimsObject(group.resource_claims ?? {}),
          browser_stage: source.browser_stage,
          browser_session_group: sessionGroup,
          browser_group: clone(group),
          env: browserGroupWorkerEnvFromPlan(browserWorkerSlotPlan, group),
          command: command("browser_group", {
            service_target: scheduleTarget,
            browser_stage: source.browser_stage,
            group_id: group.id,
          }),
          order: sourceIndex,
        });
      }
      continue;
    }

    if (source.type === "make_target") {
      counted.push({
        id: source.target,
        kind: "make_target",
        class: source.class,
        target: source.target,
        label: source.target,
        aggregate_target: source.target,
        priority: priority(source.priority),
        weight_ms: source.weight_ms,
        needs: source.needs ?? [],
        completion_keys: [source.target],
        failure_keys: [source.target],
        resource_claims: resourceClaimsObject(source.resource_claims ?? {}),
        command: command("make_target", { target: source.target }),
        order: sourceIndex,
      });
      continue;
    }

    if (source.type !== "go_shards") {
      throw new Error(`${scheduleTarget} source ${source.target} has unsupported type ${source.type}`);
    }

    const shards = collectGoShardsForTarget(repoRoot, source.target, {
      defaultCheckOnly: source.default_check_required === true,
    });
    if (shards.length === 0) {
      throw new Error(`${scheduleTarget} go_shards source ${source.target} selected no shards`);
    }
    aggregate.push({
      id: `finalize:${source.target}`,
      kind: "aggregate_finalize",
      class: source.class,
      target: source.target,
      label: `finalize/${source.target}`,
      aggregate_target: source.target,
      priority: priority(source.priority),
      weight_ms: 1,
      needs: shards.map((shard) => shardCompletionKey(shard.name)),
      completion_keys: [source.target],
      failure_keys: [source.target],
      count_in_total: false,
      counts_started: false,
      resource_claims: {},
      shard_names: shards.map((shard) => shard.name),
      unblock_label: source.target,
      command: command("go_shard_finalize", {
        target: source.target,
        service_target: scheduleTarget,
      }),
      order: sourceIndex,
    });
    for (const shard of shards) {
      if (shardWorkByName.has(shard.name)) {
        continue;
      }
      const runtime = shardRuntimeConfig(source, shard);
      addDirectRuntimeProducerUnits(runtimeProducerUnitsByTarget, runtime, source, sourceIndex);
      const env = mergeEnv(source.env, runtime.env);
      const unit = {
        id: `${source.target}:${shard.name}`,
        kind: "go_shard",
        class: source.class,
        target: source.target,
        label: `${source.target}/${shard.name}`,
        aggregate_target: source.target,
        priority: priority(source.priority),
        weight_ms: shard.weight_ms,
        needs: sortedUnique([...(source.needs ?? []), ...runtime.needs]),
        completion_keys: [shardCompletionKey(shard.name)],
        failure_keys: [shardCompletionKey(shard.name)],
        running_dependency_keys: [source.target],
        complete_on_failure: true,
        shard: shard.name,
        scheduler_profile: shard.scheduler_profile,
        resource_claims: mergeClaims(
          source.resource_claims,
          schedulerClaimsForShard(shard),
        ),
        ...(Object.keys(env).length > 0 ? { env } : {}),
        ...(runtime.runtimeBinaries.length > 0 ? { runtime_binaries: runtime.runtimeBinaries } : {}),
        command: command("go_shard", {
          target: source.target,
          shard: shard.name,
          service_target: scheduleTarget,
        }),
        order: sourceIndex,
      };
      shardWorkByName.set(shard.name, unit);
      counted.push(unit);
    }
  }

  for (const info of sessionInfos.values()) {
    if (!sharedBrowserSession(info)) {
      continue;
    }
    const finalizerTarget = browserSessionFinalizerTarget(info.group);
    aggregate.push({
      id: `browser-session-finalizer:${info.group}`,
      kind: "browser_session_finalizer",
      class: info.firstSource?.class,
      target: finalizerTarget,
      label: `${info.group}/session-finalizer`,
      aggregate_target: finalizerTarget,
      priority: info.priority,
      weight_ms: 1,
      needs: info.groupNeeds,
      completion_keys: [browserSessionFinalizerCompletionKey(info.group)],
      failure_keys: [browserSessionFinalizerCompletionKey(info.group)],
      count_in_total: false,
      counts_started: false,
      resource_claims: {},
      release_retained_resource_claims: info.retainedClaims,
      browser_session_group: info.group,
      ...(info.isolationReason
        ? { browser_session_isolation_reason: info.isolationReason }
        : {}),
      command: command("browser_session_finalizer", {
        service_target: scheduleTarget,
        browser_session_group: info.group,
      }),
      order: info.finalizerIndex,
    });
  }

  const workUnits = [
    ...[...runtimeProducerUnitsByTarget.values()],
    ...counted.sort(
      (left, right) =>
        (right.priority ?? 0) - (left.priority ?? 0) ||
        right.weight_ms - left.weight_ms ||
        (left.order ?? 0) - (right.order ?? 0) ||
        left.id.localeCompare(right.id),
    ),
    ...aggregate,
  ];
  return workUnits.map(({ order: _order, ...unit }) => clone(unit));
}
