import {
  goShardSchedulerProfileClaims,
  mapServiceBackedClaimsToCheckClaims,
} from "../../scheduler/scheduler-resource-policy.mjs";
import { addClaim, resourceClaimsObject } from "./schedule-utils.mjs";

export function sourceClaimsForShard(source, shard) {
  return mergeClaims(
    source.resource_claims,
    source.resource_claims_by_execution_family?.[shard.aggregate_name],
  );
}

export function checkClaimsForShard(source, shard) {
  const claims = new Map(Object.entries(mapServiceBackedClaimsToCheckClaims(sourceClaimsForShard(source, shard))));
  for (const [resource, amount] of Object.entries(
    goShardSchedulerProfileClaims(shard.scheduler_profile, { scheduler: "check" }),
  )) {
    addClaim(claims, resource, amount);
  }
  return resourceClaimsObject(Object.fromEntries(claims.entries()));
}

export function schedulerClaimsForShard(shard) {
  return goShardSchedulerProfileClaims(shard.scheduler_profile, {
    scheduler: "service_backed",
  });
}

export function mergeClaims(left, ...claimObjects) {
  const claims = new Map(Object.entries(left ?? {}));
  for (const claimObject of claimObjects) {
    for (const [resource, amount] of Object.entries(claimObject ?? {})) {
      addClaim(claims, resource, amount);
    }
  }
  return resourceClaimsObject(Object.fromEntries(claims.entries()));
}

export function browserGroupClaims(rawClaims) {
  const claims = new Map(Object.entries(
    mapServiceBackedClaimsToCheckClaims(rawClaims ?? {}, { ensureHost: true }),
  ));
  if (claims.has("process") && claims.get("process") !== 1) {
    throw new Error("browser group resource claims must declare process=1");
  }
  claims.set("process", 1);
  return resourceClaimsObject(Object.fromEntries(claims.entries()));
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

export function retainedBrowserStageClaims(rawClaims) {
  const mapped = mapServiceBackedClaimsToCheckClaims(rawClaims, { ensureHost: true });
  return retainedBrowserStageClaimsFromEntries(Object.entries(mapped));
}

export function directRetainedBrowserStageClaims(rawClaims) {
  return retainedBrowserStageClaimsFromEntries(Object.entries(rawClaims ?? {}));
}

export function directRuntimeProducerClaims({ scheduler = "service_backed" } = {}) {
  return goShardSchedulerProfileClaims("balanced", {
    scheduler,
  });
}
