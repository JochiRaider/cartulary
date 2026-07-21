export const schedulerFamilyValues = Object.freeze([
  "check",
  "sequence",
  "service_backed",
  "test_slice",
]);

export const schedulerFamilySet = new Set(schedulerFamilyValues);

export const schedulerCapacityProfilesByFamily = Object.freeze({
  check: Object.freeze(["check_default"]),
  sequence: Object.freeze(["sequence_adaptive"]),
  service_backed: Object.freeze([
    "service_backed_full",
    "service_backed_backend",
  ]),
  test_slice: Object.freeze(["test_slice_default"]),
});

export const schedulerCapacityProfileValues = Object.freeze(
  schedulerFamilyValues.flatMap(
    (family) => schedulerCapacityProfilesByFamily[family] ?? [],
  ),
);

const schedulerCapacityProfileSet = new Set(
  schedulerCapacityProfileValues,
);

const schedulerAutoPolicyValues = Object.freeze([
  "host_cpu",
  "host_io",
  "host_process_slots",
  "service_backed_go_cpu",
  "service_backed_go_io",
  "service_backed_browser_stack",
  "service_backed_postgres_reset",
  "service_backed_postgres_clone",
]);

const schedulerAutoPolicySet = new Set(schedulerAutoPolicyValues);

export function isSchedulerFamily(value) {
  return schedulerFamilySet.has(value);
}

export function requireSchedulerFamily(value, label) {
  if (typeof value !== "string" || !isSchedulerFamily(value)) {
    throw new Error(`${label} must be one of ${schedulerFamilyValues.join("|")}`);
  }
  return value;
}

function isSchedulerCapacityProfile(value) {
  return schedulerCapacityProfileSet.has(value);
}

function requireSchedulerCapacityProfile(value, label) {
  if (typeof value !== "string" || !isSchedulerCapacityProfile(value)) {
    throw new Error(
      `${label} must be one of ${schedulerCapacityProfileValues.join("|")}`,
    );
  }
  return value;
}

export function schedulerFamilyForCapacityProfile(profile) {
  for (const [family, profiles] of Object.entries(
    schedulerCapacityProfilesByFamily,
  )) {
    if (profiles.includes(profile)) {
      return family;
    }
  }
  return null;
}

export function requireSchedulerCapacityProfileForFamily(profile, family, label) {
  const schedulerKind = requireSchedulerFamily(family, `${label}.scheduler`);
  const profileName = requireSchedulerCapacityProfile(
    profile,
    `${label}.capacity_profile`,
  );
  const expectedFamily = schedulerFamilyForCapacityProfile(profileName);
  if (expectedFamily !== schedulerKind) {
    throw new Error(
      `${label}.capacity_profile ${profileName} is not valid for ${schedulerKind} scheduler`,
    );
  }
  return profileName;
}

function isSchedulerAutoPolicy(value) {
  return schedulerAutoPolicySet.has(value);
}

export function requireSchedulerAutoPolicy(value, label) {
  if (typeof value !== "string" || !isSchedulerAutoPolicy(value)) {
    throw new Error(
      `${label} must be one of ${schedulerAutoPolicyValues.join("|")}`,
    );
  }
  return value;
}
