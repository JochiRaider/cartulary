import {
  browserGroupClaims,
  checkClaimsForShard,
  mapServiceBackedClaimsToCheckClaims as mapServiceBackedClaimsToCheckClaimsFromPolicy,
  mergeClaims,
  schedulerClaimsForShard,
} from "./schedule-resource-claims.mjs";
import {
  browserGroupCompletionKey,
  browserGroupNeeds,
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

const serviceSessionResource = "suite_service_stack";
const buildServerTarget = "build-server";
const buildMigrateTarget = "build-migrate";
const buildWebTarget = "build-web";
const testServiceImagesTarget = "test-service-images";

export function mapServiceBackedClaimsToCheckClaims(rawClaims, { ensureHost = false } = {}) {
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
  const sessionInfos = browserSessionInfos(serviceSchedule.work_unit_sources ?? [], {
    serviceSessionKey,
    parentNeeds,
    sourceNeeds,
    browserStageExtraNeeds,
    priority,
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
        make_prerequisite_policy: "skip",
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
        make_prerequisite_policy: "skip",
        resource_claims: resourceClaimsObject(source.resource_claims ?? {}),
        command: command("make_target", { target: source.target }),
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
