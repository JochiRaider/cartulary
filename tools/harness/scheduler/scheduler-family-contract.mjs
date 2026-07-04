export const schedulerFamilyValues = Object.freeze([
  "check",
  "service_backed",
  "phase_slice",
]);

export const schedulerFamilySet = new Set(schedulerFamilyValues);

export function isSchedulerFamily(value) {
  return schedulerFamilySet.has(value);
}

export function requireSchedulerFamily(value, label) {
  if (typeof value !== "string" || !isSchedulerFamily(value)) {
    throw new Error(`${label} must be one of ${schedulerFamilyValues.join("|")}`);
  }
  return value;
}
