import {
  browserGroupCompletionKey,
  browserGroupNeeds,
  browserGroupWorkerEnvFromPlan,
  browserGroupWorkerSlotPlan,
  browserStageCompletionNeeds,
  browserStageSessionKey,
} from "../../scheduler/adapters/browser.mjs";
import {
  directRetainedBrowserStageClaims,
  retainedBrowserStageClaims,
} from "./schedule-resource-claims.mjs";
import { mapServiceBackedClaimsToCheckClaims } from "../../scheduler/scheduler-resource-policy.mjs";
import { resourceClaimsObject, unionStrings } from "./schedule-utils.mjs";

export {
  browserGroupCompletionKey,
  browserGroupNeeds,
  browserGroupWorkerEnvFromPlan,
  browserGroupWorkerSlotPlan,
  browserStageCompletionNeeds,
  browserStageSessionKey,
};

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

export function browserSessionGroupName(source) {
  const raw = source.browser_session_group;
  if (typeof raw === "string" && raw.trim() !== "") {
    return raw.trim();
  }
  return source.target;
}

function browserSessionIsolationReason(source) {
  const raw = source.browser_session_isolation_reason;
  return typeof raw === "string" && raw.trim() !== "" ? raw.trim() : "";
}

export function browserGroupSessionGroupName(source, group) {
  const raw = group?.browser_session_group;
  if (typeof raw === "string" && raw.trim() !== "") {
    return raw.trim();
  }
  return browserSessionGroupName(source);
}

function browserGroupSessionIsolationReason(source, group) {
  const raw = group?.browser_session_isolation_reason;
  if (typeof raw === "string" && raw.trim() !== "") {
    return raw.trim();
  }
  return browserSessionIsolationReason(source);
}

export function browserSessionRetainsOwnTarget(source) {
  return browserSessionGroupName(source) === source.target;
}

export function browserSessionFinalizerTarget(sessionGroup) {
  return `browser-session-${sessionGroup.replaceAll(/[^A-Za-z0-9_.-]/g, "-")}`;
}

export function browserSessionFinalizerCompletionKey(sessionGroup) {
  return `browser_session_finalizer:${sessionGroup}`;
}

export function sharedBrowserSession(info) {
  return (info?.sources?.length ?? 0) > 1;
}

export function browserSessionWorkUnitID(scheduleTarget, source) {
  const group = browserSessionGroupName(source);
  if (browserSessionRetainsOwnTarget(source)) {
    return `${scheduleTarget}:browser-stage-session:${source.browser_stage}`;
  }
  return `${scheduleTarget}:browser-stage-session:${group}`;
}

function emptySessionInfo(group, source, sourceIndex) {
  return {
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
    groups: [],
  };
}

function applySessionGroup(info, source, group, {
  sourceClaims,
  retainedClaims,
  needs,
  priority,
}) {
  if (!info.sources.includes(source)) {
    info.sources.push(source);
  }
  info.groups.push(group);
  info.resourceClaims = mergeSessionClaims(info.resourceClaims, sourceClaims);
  info.retainedClaims = mergeSessionClaims(info.retainedClaims, retainedClaims);
  info.needs = unionStrings(info.needs, needs);
  info.groupNeeds = unionStrings(info.groupNeeds, browserStageCompletionNeeds([group]));
  info.priority = Math.max(info.priority, priority);
  info.weightMs = Math.max(info.weightMs, source.weight_ms);
  const isolationReason = browserGroupSessionIsolationReason(source, group);
  if (isolationReason) {
    info.isolationReason = isolationReason;
  }
}

export function browserSessionInfos(sources, {
  serviceSessionKey,
  parentNeeds,
  sourceNeeds,
  browserStageExtraNeeds,
  priority,
}) {
  const infos = new Map();
  for (const [sourceIndex, source] of sources.entries()) {
    if (source.type !== "browser_stage") {
      continue;
    }
    for (const browserGroup of source.groups ?? []) {
      const group = browserGroupSessionGroupName(source, browserGroup);
      const info = infos.get(group) ?? emptySessionInfo(group, source, sourceIndex);
      applySessionGroup(info, source, browserGroup, {
        sourceClaims: mapServiceBackedClaimsToCheckClaims(source.resource_claims, {
          ensureHost: true,
        }),
        retainedClaims: retainedBrowserStageClaims(source.resource_claims),
        needs: sourceNeeds(source, serviceSessionKey, browserStageExtraNeeds(parentNeeds)),
        priority: priority(source.priority),
      });
      infos.set(group, info);
    }
  }
  return infos;
}

export function directBrowserSessionInfos(sources, { priority }) {
  const infos = new Map();
  for (const [sourceIndex, source] of sources.entries()) {
    if (source.type !== "browser_stage") {
      continue;
    }
    for (const browserGroup of source.groups ?? []) {
      const group = browserGroupSessionGroupName(source, browserGroup);
      const info = infos.get(group) ?? emptySessionInfo(group, source, sourceIndex);
      applySessionGroup(info, source, browserGroup, {
        sourceClaims: resourceClaimsObject(source.resource_claims ?? {}),
        retainedClaims: directRetainedBrowserStageClaims(source.resource_claims),
        needs: source.needs ?? [],
        priority: priority(source.priority),
      });
      infos.set(group, info);
    }
  }
  return infos;
}
