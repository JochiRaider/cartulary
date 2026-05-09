import {
  browserGroupCompletionKey,
  browserGroupNeeds,
  browserGroupWorkerEnv,
  browserStageCompletionNeeds,
  browserStageSessionKey,
} from "./browser-scheduler-dependencies.mjs";
import { collectGoShardsForTarget } from "./go-shard-plan.mjs";

const serviceSessionResource = "suite_service_stack";
const goCPUResource = "go_cpu";
const goIOResource = "go_io";
const hostCPUResource = "host_cpu";
const hostIOResource = "host_io";
const postgresResetResource = "postgres_reset";
const buildServerTarget = "build-server";
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

function checkClaimsForShard(sourceClaims, shard) {
  const claims = new Map(Object.entries(mapServiceBackedClaimsToCheckClaims(sourceClaims)));
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
      addClaim(claims, hostIOResource, 3);
      addClaim(claims, postgresResetResource, 1);
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

function serviceSessionNeeds(parentNeeds) {
  return parentNeeds.filter((need) => need !== buildServerTarget);
}

function browserStageExtraNeeds(parentNeeds) {
  return parentNeeds.includes(buildServerTarget) ? [buildServerTarget] : [];
}

function retainedBrowserStageClaims(rawClaims) {
  const mapped = mapServiceBackedClaimsToCheckClaims(rawClaims, { ensureHost: true });
  return resourceClaimsObject(
    Object.fromEntries(
      Object.entries(mapped).filter(([resource]) => resource !== hostCPUResource && resource !== hostIOResource),
    ),
  );
}

function browserGroupClaims(rawClaims) {
  return mapServiceBackedClaimsToCheckClaims(rawClaims ?? {}, { ensureHost: true });
}

function priority(value) {
  return Number.isInteger(value) && value > 0 ? value : defaultSchedulerPriority;
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
  const expanded = [
    {
      id: `${scheduleTarget}:service-session`,
      kind: "service_session",
      target: scheduleTarget,
      label: `${scheduleTarget}/service-session`,
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
    },
  ];

  for (const [sourceIndex, source] of (serviceSchedule.work_unit_sources ?? []).entries()) {
    if (source.type === "browser_stage") {
      const retainedClaims = retainedBrowserStageClaims(source.resource_claims);
      const stageSessionKey = browserStageSessionKey(source.target);
      expanded.push({
        id: `${scheduleTarget}:browser-stage-session:${source.browser_stage}`,
        kind: "browser_stage_session",
        target: source.target,
        label: `${source.target}/stage-session`,
        aggregate_target: source.target,
        priority: priority(source.priority),
        weight_ms: source.weight_ms,
        needs: sourceNeeds(source, serviceSessionKey, browserStageExtraNeeds(parentNeeds)),
        completion_keys: [stageSessionKey],
        failure_keys: [stageSessionKey],
        resource_claims: mapServiceBackedClaimsToCheckClaims(source.resource_claims, {
          ensureHost: true,
        }),
        retained_resource_claims: retainedClaims,
        service_session: {
          target: scheduleTarget,
        },
        browser_stage: source.browser_stage,
        order: sourceIndex,
      });
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
        release_retained_resource_claims: retainedClaims,
        service_session: {
          target: scheduleTarget,
        },
        browser_stage: source.browser_stage,
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
          needs: browserGroupNeeds(stageSessionKey),
          completion_keys: [browserGroupCompletionKey(group.id)],
          failure_keys: [browserGroupCompletionKey(group.id)],
          resource_claims: browserGroupClaims(group.resource_claims),
          service_session: {
            target: scheduleTarget,
          },
          browser_stage: source.browser_stage,
          browser_group: clone(group),
          env: browserGroupWorkerEnv(source.groups, group),
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
        completion_keys: [source.target],
        failure_keys: [source.target],
        resource_claims: mapServiceBackedClaimsToCheckClaims(source.resource_claims, {
          ensureHost: true,
        }),
        service_session: {
          target: scheduleTarget,
        },
        order: sourceIndex,
      });
      continue;
    }

    if (source.type !== "go_shards") {
      throw new Error(`${scheduleTarget} source ${source.target} has unsupported type ${source.type}`);
    }

    const shards = collectGoShardsForTarget(repoRoot, source.target);
    if (shards.length === 0) {
      throw new Error(`${scheduleTarget} go_shards source ${source.target} selected no shards`);
    }
    expanded.push({
      id: `${scheduleTarget}:finalize:${source.target}`,
      kind: "finalizer",
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
      order: sourceIndex,
    });
    for (const shard of shards) {
      expanded.push({
        id: `${scheduleTarget}:${source.target}:${shard.name}`,
        kind: "go_shard",
        target: source.target,
        label: `${source.target}/${shard.name}`,
        aggregate_target: source.target,
        priority: priority(source.priority),
        weight_ms: shard.weight_ms,
        needs: sourceNeeds(source, serviceSessionKey),
        completion_keys: [shardCompletionKey(shard.name)],
        failure_keys: [shardCompletionKey(shard.name)],
        running_dependency_keys: [source.target],
        complete_on_failure: true,
        shard: shard.name,
        scheduler_profile: shard.scheduler_profile,
        resource_claims: checkClaimsForShard(source.resource_claims, shard),
        service_session: {
          target: scheduleTarget,
        },
        order: sourceIndex,
      });
    }
  }

  expanded.push({
    id: `${scheduleTarget}:complete`,
    kind: "service_complete",
    target: scheduleTarget,
    label: `${scheduleTarget}/complete`,
    weight_ms: 1,
    needs: (serviceSchedule.work_unit_sources ?? []).map((source) => source.target),
    completion_keys: [scheduleTarget],
    failure_keys: [scheduleTarget],
    produces_summary_targets: parentUnit.produces_summary_targets ?? [scheduleTarget],
    count_in_total: false,
    counts_started: false,
    resource_claims: {},
    service_session: {
      target: scheduleTarget,
    },
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
