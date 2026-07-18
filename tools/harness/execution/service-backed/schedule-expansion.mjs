import {
  browserGroupClaims,
  checkClaimsForShard,
  mergeClaims,
  schedulerClaimsForShard,
  sourceClaimsForShard,
} from "./schedule-resource-claims.mjs";
import { mapServiceBackedClaimsToCheckClaims as mapServiceBackedClaimsToCheckClaimsFromPolicy } from "../../scheduler/scheduler-resource-policy.mjs";
import {
  browserGroupCompletionKey,
  browserGroupNeeds,
  browserGroupSessionGroupName,
  browserGroupWorkerEnvFromPlan,
  browserGroupWorkerSlotPlan,
  browserSessionFinalizerCompletionKey,
  browserSessionFinalizerTarget,
  browserSessionGroupName,
  browserSessionInfos,
  browserSessionRetainsOwnTarget,
  browserSessionWorkUnitID,
  browserStageCompletionNeeds,
  browserStageSessionKey,
  directBrowserSessionInfos,
  sharedBrowserSession,
} from "./schedule-browser-planning.mjs";
import {
  addDirectRuntimeProducerUnits,
  collectServiceBackedGoShards,
  shardCompletionKey,
  shardRuntimeConfig,
} from "./schedule-go-planning.mjs";
import {
  clone,
  command,
  mergeEnv,
  priority,
  resourceClaimsObject,
  sortedUnique,
} from "./schedule-utils.mjs";
import { readinessAttributionForMakeTarget } from "../../scheduler/scheduler-manifest.mjs";

const serviceSessionResource = "suite_service_stack";
const buildServerTarget = "build-server";
const buildMigrateTarget = "build-migrate";
const buildWebTarget = "build-web";
const testServiceImagesTarget = "test-service-images";

function sessionInfosForSource(sessionInfos, source) {
  return Array.from(sessionInfos.values()).filter((info) => info.sources.includes(source));
}

function sessionInfoForBrowserGroup(sessionInfos, source, group) {
  const sessionGroup = browserGroupSessionGroupName(source, group);
  const info = sessionInfos.get(sessionGroup);
  if (!info) {
    throw new Error(`browser group ${group.id} has no session info for ${sessionGroup}`);
  }
  return info;
}

function separateSessionFinalizerNeeded(info, sessionInfos) {
  if (sharedBrowserSession(info)) {
    return true;
  }
  return info.sources.some((source) => sessionInfosForSource(sessionInfos, source).length > 1);
}

function stageCompleteSessionInfo(source, sessionInfos) {
  const infos = sessionInfosForSource(sessionInfos, source);
  return infos[0] ?? null;
}

function browserSessionWorkUnitIDForInfo(scheduleTarget, source, info) {
  if (info.group === browserSessionGroupName(source)) {
    return browserSessionWorkUnitID(scheduleTarget, source);
  }
  return `${scheduleTarget}:browser-stage-session:${info.group}`;
}

function infoSessionRetainsSourceTarget(source, info) {
  return info.group === browserSessionGroupName(source) && browserSessionRetainsOwnTarget(source);
}

function browserGroupSelectionEnv(group) {
  const env = {
    CARTULARY_BROWSER_RUNTIME_PROFILE_ID:
      String(group.runtime_profile_id ?? "default").trim() || "default",
  };
  if (Array.isArray(group.selected_row_ids) && group.selected_row_ids.length > 0) {
    env.CARTULARY_BROWSER_SELECTED_ROW_IDS = group.selected_row_ids.join(",");
  }
  if (typeof group.browser_session_group === "string" && group.browser_session_group.trim() !== "") {
    env.CARTULARY_BROWSER_SESSION_GROUP = group.browser_session_group.trim();
  }
  return env;
}

function browserGroupEnvFromPlan(browserWorkerSlotPlan, group) {
  return mergeEnv(
    browserGroupWorkerEnvFromPlan(browserWorkerSlotPlan, group),
    browserGroupSelectionEnv(group),
    group.env,
  );
}

function scheduledBrowserGroupNeeds(previousStatefulGroupBySession, group, sessionInfo) {
  const sessionKey = sessionInfo.sessionKey ?? browserStageSessionKey(group.target);
  if (group.kind !== "stateful_partition") {
    return browserGroupNeeds(sessionKey);
  }
  const previousGroupCompletionKey = previousStatefulGroupBySession.get(sessionInfo.group) ?? null;
  const needs = browserGroupNeeds(sessionKey, previousGroupCompletionKey);
  previousStatefulGroupBySession.set(sessionInfo.group, browserGroupCompletionKey(group.id));
  return needs;
}

function mapServiceBackedClaimsToCheckClaims(rawClaims, { ensureHost = false } = {}) {
  return mapServiceBackedClaimsToCheckClaimsFromPolicy(rawClaims, { ensureHost });
}

function sourceNeeds(source, serviceSessionKey, extraNeeds = []) {
  return [serviceSessionKey, ...extraNeeds, ...(source.needs ?? [])];
}

function serviceSessionNeeds(parentNeeds) {
  return parentNeeds.filter((need) => need === testServiceImagesTarget);
}

function browserStageExtraNeeds(parentNeeds) {
  return [buildWebTarget, buildServerTarget, buildMigrateTarget].filter((need) =>
    parentNeeds.includes(need),
  );
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
  const runtimeProducerUnitsByTarget = new Map();
  const sessionInfos = browserSessionInfos(serviceSchedule.work_unit_sources ?? [], {
    serviceSessionKey,
    parentNeeds,
    sourceNeeds,
    browserStageExtraNeeds,
    priority,
  });
  const browserWorkerSlotPlan = browserGroupWorkerSlotPlan(serviceSchedule.work_unit_sources ?? []);
  const previousStatefulGroupBySession = new Map();
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
      const sourceSessionInfos = sessionInfosForSource(sessionInfos, source);
      for (const sessionInfo of sourceSessionInfos.filter((info) => info.firstSource === source)) {
        const sessionGroup = sessionInfo.group;
        expanded.push({
          id: browserSessionWorkUnitIDForInfo(scheduleTarget, source, sessionInfo),
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
          env: {
            CARTULARY_BROWSER_RUNTIME_PROFILE_ID:
              sessionInfo.runtimeProfileID || "default",
          },
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
      const completeSessionInfo = stageCompleteSessionInfo(source, sessionInfos);
      const completeSessionGroup = completeSessionInfo?.group ?? browserSessionGroupName(source);
      const finalizeOnStageComplete =
        sourceSessionInfos.length === 1 &&
        completeSessionInfo?.finalizerSource === source &&
        !separateSessionFinalizerNeeded(completeSessionInfo, sessionInfos);
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
          ? (completeSessionInfo?.retainedClaims ?? {})
          : {},
        service_session: {
          target: scheduleTarget,
        },
        browser_stage: source.browser_stage,
        browser_session_group: completeSessionGroup,
        env: {
          ...(source.env ?? {}),
          CARTULARY_BROWSER_RUNTIME_PROFILE_ID:
            completeSessionInfo?.runtimeProfileID || "default",
        },
        browser_session_finalizer: finalizeOnStageComplete,
        ...(completeSessionInfo?.isolationReason
          ? { browser_session_isolation_reason: completeSessionInfo.isolationReason }
          : {}),
        command: command("browser_stage_complete", {
          service_target: scheduleTarget,
          browser_stage: source.browser_stage,
        }),
        order: sourceIndex,
      });
      for (const group of source.groups ?? []) {
        const groupSessionInfo = sessionInfoForBrowserGroup(sessionInfos, source, group);
        expanded.push({
          id: `${scheduleTarget}:${group.id}`,
          kind: "browser_group",
          target: group.target,
          label: `${source.target}/${group.name}`,
          aggregate_target: source.target,
          priority: priority(group.priority ?? source.priority),
          weight_ms: group.weight_ms,
          needs: scheduledBrowserGroupNeeds(
            previousStatefulGroupBySession,
            group,
            groupSessionInfo,
          ),
          completion_keys: [browserGroupCompletionKey(group.id)],
          failure_keys: [browserGroupCompletionKey(group.id)],
          resource_claims: browserGroupClaims(group.resource_claims),
          service_session: {
            target: scheduleTarget,
          },
          browser_stage: source.browser_stage,
          browser_session_group: groupSessionInfo.group,
          browser_group: clone(group),
          env: browserGroupEnvFromPlan(browserWorkerSlotPlan, group),
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
      const readinessAttribution = readinessAttributionForMakeTarget(source.target);
      expanded.push({
        id: `${scheduleTarget}:${source.target}`,
        kind: "service_make_target",
        target: source.target,
        label: source.target,
        aggregate_target: source.target,
        priority: priority(source.priority),
        weight_ms: source.weight_ms,
        needs: sourceNeeds(source, serviceSessionKey),
        ...(source.runtime_binaries ? { runtime_binaries: clone(source.runtime_binaries) } : {}),
        completion_keys: [source.target],
        failure_keys: [source.target],
        make_prerequisite_policy: "skip",
        resource_claims: mapServiceBackedClaimsToCheckClaims(source.resource_claims, {
          ensureHost: true,
        }),
        service_session: {
          target: scheduleTarget,
        },
        command: command("make_target", { target: source.target, service_target: scheduleTarget }),
        ...(readinessAttribution ? { readiness_attribution: readinessAttribution } : {}),
        order: sourceIndex,
      });
      continue;
    }

    if (source.type !== "go_shards") {
      throw new Error(`${scheduleTarget} source ${source.target} has unsupported type ${source.type}`);
    }

    const shards = collectServiceBackedGoShards(repoRoot, source, scheduleTarget);
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
      addDirectRuntimeProducerUnits(
        runtimeProducerUnitsByTarget,
        runtime,
        source,
        sourceIndex,
        priority,
        {
          scheduler: "check",
          // The parent check unit already owns its declared build prerequisites.
          // Only emit producers introduced by a runtime-binary requirement (such
          // as server-harness) so completion keys stay unique across the schedule.
          omitTargets: new Set(parentNeeds),
        },
      );
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
    if (!separateSessionFinalizerNeeded(info, sessionInfos)) {
      continue;
    }
    const finalizerKey = browserSessionFinalizerCompletionKey(info.group);
    const finalizerTarget = info.firstSource?.target ?? browserSessionFinalizerTarget(info.group);
    sharedSessionFinalizerKeys.push(finalizerKey);
    expanded.push({
      id: `${scheduleTarget}:browser-session-finalizer:${info.group}`,
      kind: "browser_session_finalizer",
      target: finalizerTarget,
      label: `${info.group}/session-finalizer`,
      aggregate_target: finalizerTarget,
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
      browser_stage: info.firstSource?.browser_stage,
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

  return [...runtimeProducerUnitsByTarget.values(), ...expanded]
    .sort(
      (left, right) =>
        (right.priority ?? 0) - (left.priority ?? 0) ||
        right.weight_ms - left.weight_ms ||
        (left.order ?? 0) - (right.order ?? 0) ||
        left.id.localeCompare(right.id),
    )
    .map(({ order: _order, ...unit }) => clone(unit));
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
  const sessionInfos = directBrowserSessionInfos(serviceSchedule.work_unit_sources ?? [], {
    priority,
  });
  const browserWorkerSlotPlan = browserGroupWorkerSlotPlan(serviceSchedule.work_unit_sources ?? []);
  const previousStatefulGroupBySession = new Map();

  for (const [sourceIndex, source] of (serviceSchedule.work_unit_sources ?? []).entries()) {
    if (source.type === "browser_stage") {
      const sourceSessionInfos = sessionInfosForSource(sessionInfos, source);
      for (const sessionInfo of sourceSessionInfos.filter((info) => info.firstSource === source)) {
        const sessionGroup = sessionInfo.group;
        counted.push({
          id: infoSessionRetainsSourceTarget(source, sessionInfo)
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
          env: {
            CARTULARY_BROWSER_RUNTIME_PROFILE_ID:
              sessionInfo.runtimeProfileID || "default",
          },
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
      const completeSessionInfo = stageCompleteSessionInfo(source, sessionInfos);
      const completeSessionGroup = completeSessionInfo?.group ?? browserSessionGroupName(source);
      const finalizeOnStageComplete =
        sourceSessionInfos.length === 1 &&
        completeSessionInfo?.finalizerSource === source &&
        !separateSessionFinalizerNeeded(completeSessionInfo, sessionInfos);
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
          ? (completeSessionInfo?.retainedClaims ?? {})
          : {},
        browser_stage: source.browser_stage,
        browser_session_group: completeSessionGroup,
        env: {
          ...(source.env ?? {}),
          CARTULARY_BROWSER_RUNTIME_PROFILE_ID:
            completeSessionInfo?.runtimeProfileID || "default",
        },
        browser_session_finalizer: finalizeOnStageComplete,
        ...(completeSessionInfo?.isolationReason
          ? { browser_session_isolation_reason: completeSessionInfo.isolationReason }
          : {}),
        command: command("browser_stage_complete", {
          service_target: scheduleTarget,
          browser_stage: source.browser_stage,
        }),
        order: sourceIndex,
      });
      for (const group of source.groups ?? []) {
        const groupSessionInfo = sessionInfoForBrowserGroup(sessionInfos, source, group);
        counted.push({
          id: group.id,
          kind: "browser_group",
          class: source.class,
          target: group.target,
          label: `${source.target}/${group.name}`,
          aggregate_target: source.target,
          priority: priority(group.priority ?? source.priority),
          weight_ms: group.weight_ms,
          needs: scheduledBrowserGroupNeeds(
            previousStatefulGroupBySession,
            group,
            groupSessionInfo,
          ),
          completion_keys: [browserGroupCompletionKey(group.id)],
          failure_keys: [browserGroupCompletionKey(group.id)],
          resource_claims: resourceClaimsObject(group.resource_claims ?? {}),
          browser_stage: source.browser_stage,
          browser_session_group: groupSessionInfo.group,
          browser_group: clone(group),
          env: browserGroupEnvFromPlan(browserWorkerSlotPlan, group),
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
      const readinessAttribution = readinessAttributionForMakeTarget(source.target);
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
        make_prerequisite_policy: "skip",
        resource_claims: resourceClaimsObject(source.resource_claims ?? {}),
        command: command("make_target", { target: source.target }),
        ...(readinessAttribution ? { readiness_attribution: readinessAttribution } : {}),
        order: sourceIndex,
      });
      continue;
    }

    if (source.type !== "go_shards") {
      throw new Error(`${scheduleTarget} source ${source.target} has unsupported type ${source.type}`);
    }

    const shards = collectServiceBackedGoShards(repoRoot, source, scheduleTarget);
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
      addDirectRuntimeProducerUnits(runtimeProducerUnitsByTarget, runtime, source, sourceIndex, priority);
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
          sourceClaimsForShard(source, shard),
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
    if (!separateSessionFinalizerNeeded(info, sessionInfos)) {
      continue;
    }
    const finalizerTarget = info.firstSource?.target ?? browserSessionFinalizerTarget(info.group);
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
      browser_stage: info.firstSource?.browser_stage,
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
