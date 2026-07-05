import {
  phaseSliceDefaultCapacityProfile,
  resolveSchedulerResourceLimits,
  schedulerCapacityProfileLimits,
} from "../../scheduler/scheduler-resource-policy.mjs";

export function phaseSliceProfileResourceLimits(label) {
  return schedulerCapacityProfileLimits(
    "phase_slice",
    phaseSliceDefaultCapacityProfile,
    label,
  );
}

export function addGeneratedResourceLimit(resourceLimits, resourceLimitSources, resource, limit) {
  if (!resourceLimits.has(resource)) {
    resourceLimits.set(resource, limit);
    resourceLimitSources.set(resource, "generated");
  }
}

export function resolvePlanResourceLimits(plan) {
  const resolved = resolveSchedulerResourceLimits({
    scheduler: "phase_slice",
    resourceLimits: plan.resourceLimits,
    resourceLimitSources: plan.resourceLimitSources,
    label: `${plan.target} ${plan.phase} resource_limits`,
    workUnits: plan.workUnits,
    pruneToClaims: true,
  });
  plan.resourceLimits = resolved.resourceLimits;
  plan.resourceLimitSources = resolved.resourceLimitSources;
}
