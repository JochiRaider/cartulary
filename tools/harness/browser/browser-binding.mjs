// The single Node writer of a browser binding (TH-HARNESS Section 5.5, single
// writer): the stage, runtime and resource profiles, service requirement,
// session, preparation and stack a browser process is bound to, and the
// performance-fixture snapshot clone it uses. Its Go form is
// internal/testutil/browserbinding.
const bindingNames = Object.freeze({
  stage: "CARTULARY_BROWSER_STAGE",
  runtimeProfile: "CARTULARY_BROWSER_RUNTIME_PROFILE_ID",
  resourceProfile: "CARTULARY_BROWSER_RESOURCE_PROFILE_ID",
  serviceRequirement: "CARTULARY_BROWSER_SERVICE_REQUIREMENT",
  sessionContract: "CARTULARY_BROWSER_SESSION_CONTRACT",
  sessionGroup: "CARTULARY_BROWSER_SESSION_GROUP",
  stackKey: "CARTULARY_BROWSER_STACK_KEY",
  preparationLabel: "CARTULARY_BROWSER_PREPARATION_LABEL",
  preparationTimeoutMS: "CARTULARY_BROWSER_PREPARATION_TIMEOUT_MS",
  resetDrainTimeoutMS: "CARTULARY_BROWSER_RESET_DRAIN_TIMEOUT_MS",
  cohortLabel: "CARTULARY_BROWSER_COHORT_LABEL",
  acquisitionFile: "CARTULARY_BROWSER_ACQUISITION_FILE",
  sessionLeaseFile: "CARTULARY_WEB_E2E_SESSION_LEASE_FILE",
  privateSessionRoot: "CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT",
  fixtureProfile: "CARTULARY_FIXTURE_PROFILE_ID",
  snapshotKey: "CARTULARY_FIXTURE_SNAPSHOT_KEY",
  snapshotBuilderUnit: "CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID",
  fixtureRow: "CARTULARY_FIXTURE_ROW_ID",
  fixturePredicate: "CARTULARY_FIXTURE_PREDICATE_ID",
  cloneLease: "CARTULARY_FIXTURE_CLONE_LEASE_ID",
  cloneOrdinal: "CARTULARY_FIXTURE_CLONE_ORDINAL",
});

// browserBindingEnvironment writes only the fields that are set; every reader
// treats an empty binding value as unset.
export function browserBindingEnvironment(binding) {
  const unknown = Object.keys(binding).filter((field) => !Object.hasOwn(bindingNames, field));
  if (unknown.length > 0) throw new Error(`unknown browser binding fields: ${unknown.join(", ")}`);
  return Object.fromEntries(Object.entries(bindingNames)
    .filter(([field]) => binding[field] !== undefined && binding[field] !== null && binding[field] !== "")
    .map(([field, name]) => [name, String(binding[field])]));
}
